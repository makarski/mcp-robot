package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/makarski/mcp-robot/handler"
	"github.com/makarski/mcp-robot/io"
	"github.com/makarski/mcp-robot/spec"
	"github.com/makarski/mcp-robot/tools"
)

var o = os.Stdout

type (
	server struct {
		mu           sync.RWMutex
		toolsPerPage int
		tools        map[string]serverTool
		toolNames    []string // used for pagination
		info         spec.Info
	}

	serverTool struct {
		handler        handler.MCPHandler
		toolDefinition tools.ToolDefinition
	}
)

func NewServerBuilder(name, version string) *server {
	return newServer(name, version)
}

func (s *server) BuildHTTPServer() *HTTPServer {
	return &HTTPServer{
		server: s,
	}
}

func (s *server) BuildStdioServer() *StdioServer {
	return &StdioServer{
		server: s,
	}
}

func newServer(name, version string) *server {
	return &server{
		toolsPerPage: -1, // -1 means no pagination
		tools:        make(map[string]serverTool),
		info: spec.Info{
			Name:    name,
			Version: version,
		},
	}
}

func (s *server) okEmptyResponse(w io.RPCResponseWriter, rpcReq *spec.Request[int]) {
	response := spec.Response[int]{
		Jsonrpc: spec.JsonRPC,
		ID:      rpcReq.ID,
		Result:  map[string]any{},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		rw := io.NewResponseWriter(w, rpcReq.ID)
		rw.WriteError(
			spec.ErrorCodeInternalError,
			"failed to encode empty response",
		)
	}
}

func (s *server) initializeHandler(w io.RPCResponseWriter, rpcReq *spec.Request[int]) {
	capabilities := s.capabilities()

	response := spec.Response[int]{
		Jsonrpc: spec.JsonRPC,
		ID:      rpcReq.ID,
		Result: map[string]any{
			"protocolVersion": spec.ProtocolVersion,
			"capabilities":    capabilities,
			"serverInfo":      s.info,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		rw := io.NewResponseWriter(w, rpcReq.ID)
		rw.WriteError(
			spec.ErrorCodeInternalError,
			"failed to encode initialize response",
		)
	}
}

func (s *server) listToolsHandler(w io.RPCResponseWriter, rpcReq *spec.Request[int]) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rw := io.NewResponseWriter(w, rpcReq.ID)
	if len(s.toolNames) == 0 {
		rw.WriteResult(map[string]any{"tools": []tools.ToolDefinition{}})
		return
	}

	startPage := 0
	offset := 0
	nextPage := startPage
	end := len(s.tools)

	if cursor, ok := rpcReq.Params["cursor"]; ok {
		cursorStr, ok := cursor.(string)
		if !ok {
			rw.WriteError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("invalid 'cursor' parameter type: expected int or string, got %T", cursor),
			)
			return
		}

		cursorInt, err := strconv.Atoi(cursorStr)
		if err != nil {
			rw.WriteError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("cant convert 'cursor' parameter to int: %s", cursor),
			)
			return
		}
		startPage = cursorInt
	}

	if s.toolsPerPage > 0 {
		offset = startPage * s.toolsPerPage
		if offset+s.toolsPerPage < end {
			end = offset + s.toolsPerPage
			nextPage = startPage + 1
		}
	}

	if offset >= len(s.toolNames) {
		rw.WriteError(
			spec.ErrorCodeInvalidParams,
			fmt.Sprintf("invalid 'cursor' parameter: %d is out of range", startPage),
		)
		return
	}

	toolsList := make([]tools.ToolDefinition, 0, len(s.tools))
	for _, tool := range s.toolNames[offset:end] {
		toolsList = append(toolsList, s.tools[tool].toolDefinition)
	}

	listResult := map[string]any{
		"tools": toolsList,
	}

	if nextPage > 0 {
		listResult["nextCursor"] = fmt.Sprintf("%d", nextPage)
	}

	if err := rw.WriteResult(listResult); err != nil {
		rw.WriteError(
			spec.ErrorCodeInternalError,
			"failed to encode list response",
		)
		return
	}
}

func (s *server) WithTool(definition tools.ToolDefinition, toolHandler tools.ToolHandler) *server {
	s.mu.Lock()

	if _, ok := s.tools[definition.Name]; !ok {
		handler := toolHandler.MCPHandler(definition)
		s.tools[definition.Name] = serverTool{handler, definition}
		s.toolNames = append(s.toolNames, definition.Name)
	}

	s.mu.Unlock()

	return s
}

func (s *server) ToolsPerPage(toolsPerPage int) *server {
	s.mu.Lock()
	s.toolsPerPage = toolsPerPage
	s.mu.Unlock()

	return s
}

func (s *server) capabilities() map[string]spec.CapabilityParam {
	capabilities := make(map[string]spec.CapabilityParam)

	if len(s.tools) > 0 {
		capabilities["tools"] = spec.CapabilityParam{
			ListChanged: true,
		}
	}

	return capabilities
}

func (s *server) resolveHandler(rpcReq spec.Request[int]) (handler.MCPHandler, error) {
	switch rpcReq.Method {
	case spec.MethodNotificationsInitialized, spec.MethodPing:
		return handler.MCPHandlerFunc(s.okEmptyResponse), nil
	case spec.MethodInitialize:
		return handler.MCPHandlerFunc(s.initializeHandler), nil
	case spec.MethodToolsList:
		return handler.MCPHandlerFunc(s.listToolsHandler), nil
	case spec.MethodToolsCall:
		toolName, ok := rpcReq.Params["name"]
		if !ok {
			return nil, spec.NewProtocolError(spec.ErrorCodeInvalidParams, "missing or invalid 'name' parameter")
		}

		toolNameStr, ok := toolName.(string)
		if !ok {
			return nil, spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("invalid 'name' parameter type: expected string, got %T", toolName),
			)
		}

		s.mu.RLock()
		tool, ok := s.tools[toolNameStr]
		s.mu.RUnlock()
		if !ok {
			return nil, spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("tool not found: %s", toolNameStr),
			)
		}

		args, hasArgs := rpcReq.Params["arguments"]
		if !hasArgs && len(tool.toolDefinition.InputSchema.Required) > 0 {
			return nil, spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("missing required 'arguments' parameter for tool: %s", toolNameStr),
			)
		}

		if hasArgs {
			argsMap, ok := args.(map[string]any)
			if !ok {
				return nil, spec.NewProtocolError(
					spec.ErrorCodeInvalidParams,
					fmt.Sprintf("invalid 'arguments' parameter type: expected object, got %T", args),
				)
			}

			if err := tool.toolDefinition.ValidateArguments(argsMap); err != nil {
				return nil, err
			}
		}

		return tool.handler, nil
	default:
		return nil, spec.NewProtocolError(
			spec.ErrorCodeMethodNotFound,
			fmt.Sprintf("method not found: %s", rpcReq.Method),
		)
	}
}
