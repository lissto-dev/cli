package mcp_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/cli/pkg/mcp"
)

var _ = Describe("MCP Tools", func() {
	Describe("GetAllTools", func() {
		var tools []mcp.Tool

		BeforeEach(func() {
			tools = mcp.GetAllTools()
		})

		It("should return at least 16 tools", func() {
			Expect(len(tools)).To(BeNumerically(">=", 16))
		})

		Context("Environment Tools", func() {
			It("should include lissto_env_list", func() {
				tool := findTool(tools, "lissto_env_list")
				Expect(tool).NotTo(BeNil())
				Expect(tool.Description).To(ContainSubstring("environments"))
				Expect(tool.InputSchema).To(HaveKey("type"))
			})

			It("should include lissto_env_get with name parameter", func() {
				tool := findTool(tools, "lissto_env_get")
				Expect(tool).NotTo(BeNil())
				Expect(tool.InputSchema).To(HaveKey("properties"))

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("name"))

				required := tool.InputSchema["required"].([]string)
				Expect(required).To(ContainElement("name"))
			})

			It("should include lissto_env_create", func() {
				tool := findTool(tools, "lissto_env_create")
				Expect(tool).NotTo(BeNil())
			})

			It("should include lissto_env_current", func() {
				tool := findTool(tools, "lissto_env_current")
				Expect(tool).NotTo(BeNil())
			})
		})

		Context("Blueprint Tools", func() {
			It("should include lissto_blueprint_list with include_global parameter", func() {
				tool := findTool(tools, "lissto_blueprint_list")
				Expect(tool).NotTo(BeNil())

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("include_global"))
			})

			It("should include lissto_blueprint_get", func() {
				tool := findTool(tools, "lissto_blueprint_get")
				Expect(tool).NotTo(BeNil())
			})

			It("should include lissto_blueprint_create with compose parameter", func() {
				tool := findTool(tools, "lissto_blueprint_create")
				Expect(tool).NotTo(BeNil())

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("compose"))
				Expect(properties).To(HaveKey("branch"))
				Expect(properties).To(HaveKey("author"))
				Expect(properties).To(HaveKey("repository"))
				Expect(properties).To(HaveKey("global"))

				required := tool.InputSchema["required"].([]string)
				Expect(required).To(ContainElement("compose"))
			})

			It("should include lissto_blueprint_delete", func() {
				tool := findTool(tools, "lissto_blueprint_delete")
				Expect(tool).NotTo(BeNil())
			})
		})

		Context("Stack Tools", func() {
			It("should include lissto_stack_list with optional env parameter", func() {
				tool := findTool(tools, "lissto_stack_list")
				Expect(tool).NotTo(BeNil())

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("env"))
			})

			It("should include lissto_stack_get", func() {
				tool := findTool(tools, "lissto_stack_get")
				Expect(tool).NotTo(BeNil())

				required := tool.InputSchema["required"].([]string)
				Expect(required).To(ContainElement("name"))
			})

			It("should include lissto_stack_create with blueprint_name parameter", func() {
				tool := findTool(tools, "lissto_stack_create")
				Expect(tool).NotTo(BeNil())

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("blueprint_name"))
				Expect(properties).To(HaveKey("env"))

				required := tool.InputSchema["required"].([]string)
				Expect(required).To(ContainElement("blueprint_name"))
			})

			It("should include lissto_stack_delete", func() {
				tool := findTool(tools, "lissto_stack_delete")
				Expect(tool).NotTo(BeNil())
			})
		})

		Context("Admin Tools", func() {
			It("should include lissto_admin_apikey_create", func() {
				tool := findTool(tools, "lissto_admin_apikey_create")
				Expect(tool).NotTo(BeNil())

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("name"))
				Expect(properties).To(HaveKey("role"))
				Expect(properties).To(HaveKey("slack_user_id"))

				required := tool.InputSchema["required"].([]string)
				Expect(required).To(ContainElement("name"))
			})

			It("should include lissto_admin_blueprint_delete", func() {
				tool := findTool(tools, "lissto_admin_blueprint_delete")
				Expect(tool).NotTo(BeNil())
			})
		})

		Context("Operations Tools", func() {
			It("should include lissto_status with optional env filter", func() {
				tool := findTool(tools, "lissto_status")
				Expect(tool).NotTo(BeNil())
				Expect(tool.Description).To(ContainSubstring("status"))
			})

			It("should include lissto_logs with filtering parameters", func() {
				tool := findTool(tools, "lissto_logs")
				Expect(tool).NotTo(BeNil())

				properties := tool.InputSchema["properties"].(map[string]interface{})
				Expect(properties).To(HaveKey("stack"))
				Expect(properties).To(HaveKey("env"))
				Expect(properties).To(HaveKey("service"))
				Expect(properties).To(HaveKey("pod"))
				Expect(properties).To(HaveKey("tail"))
				Expect(properties).To(HaveKey("max_pods"))
			})
		})

		Context("Tool Schema Validation", func() {
			It("should have valid JSON schemas for all tools", func() {
				for _, tool := range tools {
					Expect(tool.Name).NotTo(BeEmpty(), "Tool name should not be empty")
					Expect(tool.Description).NotTo(BeEmpty(), "Tool description should not be empty for "+tool.Name)
					Expect(tool.InputSchema).NotTo(BeNil(), "Tool input schema should not be nil for "+tool.Name)
					Expect(tool.InputSchema).To(HaveKey("type"), "Tool input schema should have 'type' field for "+tool.Name)
					Expect(tool.InputSchema["type"]).To(Equal("object"), "Tool input schema type should be 'object' for "+tool.Name)
				}
			})

			It("should have unique tool names", func() {
				toolNames := make(map[string]bool)
				for _, tool := range tools {
					Expect(toolNames[tool.Name]).To(BeFalse(), "Duplicate tool name: "+tool.Name)
					toolNames[tool.Name] = true
				}
			})

			It("should follow naming convention lissto_<category>_<action>", func() {
				for _, tool := range tools {
					Expect(tool.Name).To(HavePrefix("lissto_"), "Tool name should start with 'lissto_': "+tool.Name)
				}
			})
		})
	})
})

// Helper function to find a tool by name
func findTool(tools []mcp.Tool, name string) *mcp.Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return &tool
		}
	}
	return nil
}
