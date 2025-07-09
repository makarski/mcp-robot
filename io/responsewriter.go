package io

import (
	"encoding/json"

	"github.com/makarski/mcp-robot/spec"
)

type (
	ResponseWriter struct {
		w  RPCResponseWriter
		id int // to do chnage to generic ID type as in request
	}

	RPCResponseWriter interface {
		Write(b []byte) (int, error)
	}
)

func NewResponseWriter(w RPCResponseWriter, id int) *ResponseWriter {
	return &ResponseWriter{
		w:  w,
		id: id,
	}
}

func (rw *ResponseWriter) WriteResult(result map[string]any) error {
	response := spec.Response[int]{
		Jsonrpc: spec.JsonRPC,
		ID:      rw.id,
		Result:  result,
	}

	return json.NewEncoder(rw.w).Encode(response)
}

func (rw *ResponseWriter) WriteError(code int, message string) error {
	errorResponse := spec.Response[int]{
		Jsonrpc: spec.JsonRPC,
		ID:      rw.id,
		Error: &spec.Error{
			Code:    code,
			Message: message,
		},
	}

	return json.NewEncoder(rw.w).Encode(errorResponse)
}
