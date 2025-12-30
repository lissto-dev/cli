package mcp_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/cli/pkg/mcp"
)

var _ = Describe("MCP Server", func() {
	var (
		stdin  *bytes.Buffer
		stdout *bytes.Buffer
		server *mcp.Server
	)

	BeforeEach(func() {
		stdin = &bytes.Buffer{}
		stdout = &bytes.Buffer{}
		var err error
		server, err = mcp.NewServer(stdin, stdout, "") // No log file for tests
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if server != nil {
			_ = server.Close()
		}
	})

	Describe("Initialize Request", func() {
		It("should return server capabilities", func() {
			// Send initialize request
			request := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
				"params":  map[string]interface{}{},
			}
			requestJSON, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			stdin.Write(requestJSON)
			stdin.Write([]byte("\n"))

			// Run server in goroutine (it will read one line and stop when stdin closes)
			errChan := make(chan error, 1)
			go func() {
				errChan <- server.Run()
			}()

			// Wait for response
			Eventually(stdout.Len).Should(BeNumerically(">", 0))

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(stdout.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			// Verify response structure
			Expect(response).To(HaveKey("jsonrpc"))
			Expect(response["jsonrpc"]).To(Equal("2.0"))
			Expect(response).To(HaveKey("id"))
			Expect(response["id"]).To(BeNumerically("==", 1))
			Expect(response).To(HaveKey("result"))

			result := response["result"].(map[string]interface{})
			Expect(result).To(HaveKey("protocolVersion"))
			Expect(result).To(HaveKey("capabilities"))
			Expect(result).To(HaveKey("serverInfo"))

			serverInfo := result["serverInfo"].(map[string]interface{})
			Expect(serverInfo["name"]).To(Equal("lissto-mcp"))
			Expect(serverInfo["version"]).To(Equal("0.1.0"))
		})
	})

	Describe("Tools List Request", func() {
		It("should return all available tools", func() {
			// Send tools/list request
			request := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"method":  "tools/list",
				"params":  map[string]interface{}{},
			}
			requestJSON, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			stdin.Write(requestJSON)
			stdin.Write([]byte("\n"))

			// Run server
			errChan := make(chan error, 1)
			go func() {
				errChan <- server.Run()
			}()

			// Wait for response
			Eventually(stdout.Len).Should(BeNumerically(">", 0))

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(stdout.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())

			// Verify response
			Expect(response).To(HaveKey("result"))
			result := response["result"].(map[string]interface{})
			Expect(result).To(HaveKey("tools"))

			tools := result["tools"].([]interface{})
			Expect(len(tools)).To(BeNumerically(">=", 16))

			// Verify required tools exist
			toolNames := make([]string, 0, len(tools))
			for _, tool := range tools {
				toolMap := tool.(map[string]interface{})
				toolNames = append(toolNames, toolMap["name"].(string))
			}

			Expect(toolNames).To(ContainElement("lissto_env_list"))
			Expect(toolNames).To(ContainElement("lissto_stack_list"))
			Expect(toolNames).To(ContainElement("lissto_blueprint_list"))
			Expect(toolNames).To(ContainElement("lissto_status"))
		})
	})

	Describe("Error Handling", func() {
		Context("when JSON-RPC version is invalid", func() {
			It("should return InvalidRequest error", func() {
				request := map[string]interface{}{
					"jsonrpc": "1.0",
					"id":      3,
					"method":  "initialize",
					"params":  map[string]interface{}{},
				}
				requestJSON, err := json.Marshal(request)
				Expect(err).NotTo(HaveOccurred())

				stdin.Write(requestJSON)
				stdin.Write([]byte("\n"))

				go func() { _ = server.Run() }()

				Eventually(stdout.Len).Should(BeNumerically(">", 0))

				var response map[string]interface{}
				err = json.Unmarshal(stdout.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				errorObj := response["error"].(map[string]interface{})
				Expect(errorObj["code"]).To(BeNumerically("==", -32600))
				Expect(errorObj["message"]).To(ContainSubstring("Invalid JSON-RPC version"))
			})
		})

		Context("when method is not found", func() {
			It("should return MethodNotFound error", func() {
				request := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      4,
					"method":  "invalid_method",
					"params":  map[string]interface{}{},
				}
				requestJSON, err := json.Marshal(request)
				Expect(err).NotTo(HaveOccurred())

				stdin.Write(requestJSON)
				stdin.Write([]byte("\n"))

				go func() { _ = server.Run() }()

				Eventually(stdout.Len).Should(BeNumerically(">", 0))

				var response map[string]interface{}
				err = json.Unmarshal(stdout.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				errorObj := response["error"].(map[string]interface{})
				Expect(errorObj["code"]).To(BeNumerically("==", -32601))
				Expect(errorObj["message"]).To(ContainSubstring("Method not found"))
			})
		})

		Context("when JSON is malformed", func() {
			It("should return ParseError", func() {
				stdin.Write([]byte("not valid json\n"))

				go func() { _ = server.Run() }()

				Eventually(stdout.Len).Should(BeNumerically(">", 0))

				var response map[string]interface{}
				err := json.Unmarshal(stdout.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveKey("error"))
				errorObj := response["error"].(map[string]interface{})
				Expect(errorObj["code"]).To(BeNumerically("==", -32700))
				Expect(errorObj["message"]).To(ContainSubstring("Parse error"))
			})
		})
	})

	Describe("Notification Handling", func() {
		Context("when receiving initialized notification", func() {
			It("should not send a response", func() {
				request := map[string]interface{}{
					"jsonrpc": "2.0",
					"method":  "initialized",
					"params":  map[string]interface{}{},
				}
				requestJSON, err := json.Marshal(request)
				Expect(err).NotTo(HaveOccurred())

				stdin.Write(requestJSON)
				stdin.Write([]byte("\n"))

				go func() { _ = server.Run() }()

				// Give it a moment to process
				Consistently(stdout.Len, "500ms").Should(Equal(0))
			})
		})
	})

	Describe("Multiple Requests", func() {
		It("should handle multiple sequential requests", func() {
			// First request
			request1 := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
				"params":  map[string]interface{}{},
			}
			req1JSON, _ := json.Marshal(request1)
			stdin.Write(req1JSON)
			stdin.Write([]byte("\n"))

			// Second request
			request2 := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"method":  "tools/list",
				"params":  map[string]interface{}{},
			}
			req2JSON, _ := json.Marshal(request2)
			stdin.Write(req2JSON)
			stdin.Write([]byte("\n"))

			go func() { _ = server.Run() }()

			// Should receive two responses
			Eventually(func() int {
				return bytes.Count(stdout.Bytes(), []byte("\n"))
			}).Should(BeNumerically(">=", 2))
		})
	})
})
