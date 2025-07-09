package tools

import (
	"fmt"

	"github.com/makarski/mcp-robot/handler"
	"github.com/makarski/mcp-robot/io"
	"github.com/makarski/mcp-robot/spec"
)

type (
	ToolFunc[TR ToolResult] func(params map[string]any) (TR, error)

	ToolHandler interface {
		MCPHandler(definition ToolDefinition) handler.MCPHandler
	}
)

func (f ToolFunc[TR]) MCPHandler(definition ToolDefinition) handler.MCPHandler {
	return handler.MCPHandlerFunc(func(w io.RPCResponseWriter, req *spec.Request[int]) {
		rw := io.NewResponseWriter(w, req.ID)

		errfmt := "failed to write response for reqID: %v: %s"
		args, ok := req.Params["arguments"].(map[string]any)
		if !ok {
			args = make(map[string]any)
		}

		result, err := f(args)
		if err != nil {
			switch e := err.(type) {
			case *spec.ProtocolError:
				rw.WriteError(e.Code, e.Message)
			default:
				writeError(rw, fmt.Sprintf(errfmt, req.ID, err))
			}
			return
		}

		if err := validateOutput(definition, result); err != nil {
			writeError(rw, fmt.Sprintf(errfmt, req.ID, err))
			return
		}

		if err := writeResult(rw, result); err != nil {
			writeError(rw, fmt.Sprintf(errfmt, req.ID, err))
			return
		}
	})
}
