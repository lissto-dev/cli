package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"` // ID must always be present per JSON-RPC 2.0 spec
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Server represents the MCP server
type Server struct {
	stdin   io.Reader
	stdout  io.Writer
	logger  *log.Logger
	logFile *os.File
}

// NewServer creates a new MCP server with optional logging
func NewServer(stdin io.Reader, stdout io.Writer, logFilePath string) (*Server, error) {
	server := &Server{
		stdin:  stdin,
		stdout: stdout,
	}

	// Setup logging if log file path is provided
	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		server.logFile = logFile
		server.logger = log.New(logFile, "", 0)
		server.log("MCP server starting at %s", time.Now().Format(time.RFC3339))
		server.log("Log file: %s", logFilePath)
	}

	return server, nil
}

// Close closes the log file if it was opened
func (s *Server) Close() error {
	if s.logFile != nil {
		s.log("MCP server shutting down at %s", time.Now().Format(time.RFC3339))
		return s.logFile.Close()
	}
	return nil
}

// log writes a log message if logging is enabled
func (s *Server) log(format string, args ...interface{}) {
	if s.logger != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		message := fmt.Sprintf(format, args...)
		s.logger.Printf("[%s] %s\n", timestamp, message)
	}
}

// Run starts the MCP server and processes requests
func (s *Server) Run() error {
	s.log("Starting to listen for requests on stdin")
	scanner := bufio.NewScanner(s.stdin)

	for scanner.Scan() {
		line := scanner.Bytes()
		s.log("Received request: %s", string(line))

		// Parse request
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.log("Parse error: %v", err)
			s.sendError(nil, ParseError, fmt.Sprintf("Parse error: %v", err), nil)
			continue
		}

		s.log("Parsed request - Method: %s, ID: %v", req.Method, req.ID)

		// Handle request
		s.handleRequest(&req)
	}

	if err := scanner.Err(); err != nil {
		s.log("Error reading stdin: %v", err)
		return fmt.Errorf("error reading stdin: %w", err)
	}

	s.log("Scanner closed, server stopping")
	return nil
}

// handleRequest processes a single JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) {
	// Check if this is a notification (no ID field)
	// Notifications must not receive any response per JSON-RPC 2.0 spec
	isNotification := req.ID == nil
	s.log("Handling request - Method: %s, IsNotification: %v", req.Method, isNotification)

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		s.log("Invalid JSON-RPC version: %s", req.JSONRPC)
		// Only send error for requests, not notifications
		if !isNotification {
			s.sendError(req.ID, InvalidRequest, "Invalid JSON-RPC version", nil)
		}
		return
	}

	// Route to appropriate handler
	switch req.Method {
	case "initialize":
		s.log("Routing to initialize handler")
		s.handleInitialize(req)
	case "initialized":
		s.log("Received initialized notification")
		// This is a notification sent by the client after initialization
		// No response required per JSON-RPC 2.0 spec
		return
	case "tools/list":
		s.log("Routing to tools/list handler")
		s.handleToolsList(req)
	case "tools/call":
		s.log("Routing to tools/call handler")
		s.handleToolsCall(req)
	default:
		s.log("Method not found: %s", req.Method)
		// Only send error for requests, not notifications
		if !isNotification {
			s.sendError(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
		}
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "lissto-mcp",
			"version": "0.1.0",
		},
	}

	s.sendResult(req.ID, result)
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *JSONRPCRequest) {
	tools := GetAllTools()
	result := map[string]interface{}{
		"tools": tools,
	}

	s.sendResult(req.ID, result)
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(req *JSONRPCRequest) {
	// Parse params
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.log("Failed to parse tool call params: %v", err)
		s.sendError(req.ID, InvalidParams, fmt.Sprintf("Invalid params: %v", err), nil)
		return
	}

	s.log("========================================")
	s.log("Tool Call Request ID: %v", req.ID)
	s.log("Tool Name: %s", params.Name)
	s.log("Tool Arguments: %+v", params.Arguments)
	s.log("========================================")

	// Execute tool with logger
	result, err := ExecuteTool(params.Name, params.Arguments, s)
	if err != nil {
		s.log("❌ TOOL EXECUTION FAILED")
		s.log("Tool: %s", params.Name)
		s.log("Error: %v", err)
		s.log("Error Type: %T", err)
		s.log("========================================")
		s.sendError(req.ID, InternalError, err.Error(), nil)
		return
	}

	// Log the result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	s.log("✅ TOOL EXECUTION SUCCESSFUL")
	s.log("Tool: %s", params.Name)
	s.log("Result: %s", string(resultJSON))
	s.log("========================================")

	// Wrap result in MCP content format
	mcpResult := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}

	s.sendResult(req.ID, mcpResult)
}

// sendResult sends a successful JSON-RPC response
func (s *Server) sendResult(id interface{}, result interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	s.sendResponse(&response)
}

// sendError sends an error JSON-RPC response
func (s *Server) sendError(id interface{}, code int, message string, _ interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    nil,
		},
	}

	s.sendResponse(&response)
}

// sendResponse writes a JSON-RPC response to stdout
func (s *Server) sendResponse(response *JSONRPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		// This should never happen, but log to stderr if it does
		s.log("Failed to marshal response: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
		return
	}

	s.log("Sending response: %s", string(data))

	// Write response followed by newline
	data = append(data, '\n')
	if _, err := s.stdout.Write(data); err != nil {
		s.log("Failed to write response: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to write response: %v\n", err)
		return
	}

	// Flush stdout to ensure response is sent immediately
	// This is critical for MCP clients like Cursor that maintain persistent connections
	if f, ok := s.stdout.(*os.File); ok {
		_ = f.Sync()
	}

	s.log("Response sent successfully")
}
