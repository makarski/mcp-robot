package mcprobot

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var o = os.Stdout

type (
	RPCResponseWriter interface {
		Write(b []byte) (int, error)
	}

	MCPHandler interface {
		ServeRPC(w RPCResponseWriter, req *Request[int])
	}

	MCPHandlerFunc func(w RPCResponseWriter, req *Request[int])
)

func (f MCPHandlerFunc) ServeRPC(w RPCResponseWriter, req *Request[int]) {
	f(w, req)
}

type (
	server struct {
		mu    sync.RWMutex
		tools map[string]serverTool
		// handlers map[string]func(Request[int]) http.Handler
		info Info
	}

	serverTool struct {
		handler        MCPHandler
		toolDefinition ToolDefinition
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
		tools: make(map[string]serverTool),
		info: Info{
			Name:    name,
			Version: version,
		},
	}
}

func (s *server) initializeHandler(w RPCResponseWriter, rpcReq *Request[int]) {
	capabilities := s.capabilities()

	response := Response[int]{
		Jsonrpc: JsonRPC,
		ID:      rpcReq.ID,
		Result: map[string]any{
			"protocolVersion": ProtocolVersion,
			"capabilities":    capabilities,
			"serverInfo":      s.info,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// http.Error(w, fmt.Sprintf("Failed to encode initialize response: %s", err), http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to encode initialize response: %s", err)
		// return
	}
}

// func (s *Server) listToolsHandler(w RPCResponseWriter, rpcReq *Request[int]) http.Handler {
func (s *server) listToolsHandler(w RPCResponseWriter, rpcReq *Request[int]) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	toolsList := make([]ToolDefinition, 0, len(s.tools))
	for _, tool := range s.tools {
		toolsList = append(toolsList, tool.toolDefinition)
	}

	response := Response[int]{
		Jsonrpc: JsonRPC,
		ID:      rpcReq.ID,
		Result: map[string]any{
			"tools": toolsList,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// todo: handle error properly
		// use the mcp specification for error handling
		// http.Error(w, fmt.Sprintf("Failed to encode tools list response: %s", err), http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to encode tools list response: %s", err)
		return
	}
}

func (s *server) WithTool(definition ToolDefinition, handler MCPHandler) *server {
	s.mu.Lock()
	s.tools[definition.Name] = serverTool{handler, definition}
	s.mu.Unlock()

	return s
}

func (s *server) capabilities() map[string]CapabilityParam {
	capabilities := make(map[string]CapabilityParam)

	if len(s.tools) > 0 {
		capabilities["tools"] = CapabilityParam{
			ListChanged: true,
		}
	}

	return capabilities
}

func (s *server) resolveHandler(rpcReq Request[int]) (MCPHandler, error) {
	switch rpcReq.Method {
	case MethodInitialize:
		return MCPHandlerFunc(s.initializeHandler), nil
	case MethodToolsList:
		return MCPHandlerFunc(s.listToolsHandler), nil
	case MethodToolsCall:
		toolName, ok := rpcReq.Params["name"]
		if !ok {
			return nil, NewProtocolError(ErrorCodeInvalidParams, "missing or invalid 'name' parameter")
		}

		toolNameStr, ok := toolName.(string)
		if !ok {
			return nil, NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("invalid 'name' parameter type: expected string, got %T", toolName),
			)
		}

		s.mu.RLock()
		tool, ok := s.tools[toolNameStr]
		s.mu.RUnlock()
		if !ok {
			return nil, NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("tool not found: %s", toolNameStr),
			)
		}

		args, hasArgs := rpcReq.Params["arguments"]
		if !hasArgs && len(tool.toolDefinition.Required) > 0 {
			return nil, NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("missing required 'arguments' parameter for tool: %s", toolNameStr),
			)
		}

		if hasArgs {
			argsMap, ok := args.(map[string]any)
			if !ok {
				return nil, NewProtocolError(
					ErrorCodeInvalidParams,
					fmt.Sprintf("invalid 'arguments' parameter type: expected object, got %T", args),
				)
			}

			if err := tool.toolDefinition.validateArguments(argsMap); err != nil {
				return nil, err
			}
		}

		return tool.handler, nil
	default:
		return nil, NewProtocolError(
			ErrorCodeMethodNotFound,
			fmt.Sprintf("method not found: %s", rpcReq.Method),
		)
	}
}
