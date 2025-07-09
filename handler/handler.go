package handler

import (
	"github.com/makarski/mcp-robot/io"
	"github.com/makarski/mcp-robot/spec"
)

type (
	MCPHandler interface {
		ServeRPC(w io.RPCResponseWriter, req *spec.Request[int])
	}

	MCPHandlerFunc func(w io.RPCResponseWriter, req *spec.Request[int])
)

func (f MCPHandlerFunc) ServeRPC(w io.RPCResponseWriter, req *spec.Request[int]) {
	f(w, req)
}
