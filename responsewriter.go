package mcprobot

import (
	"encoding/json"
	"fmt"
)

type (
	ResponseWriter struct {
		w  RPCResponseWriter
		id int // to do chnage to generic ID type as in request
	}
)

func NewResponseWriter(w RPCResponseWriter, id int) *ResponseWriter {
	return &ResponseWriter{
		w:  w,
		id: id,
	}
}

func (rw *ResponseWriter) WriteToolResult(result any) error {
	var (
		resultArray      []any
		structuredResult *ToolResultStructured
	)

	switch v := result.(type) {
	case ToolResultStructured:
		txt, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal structured tool result: %w", err)
		}

		resultArray = []any{NewToolResultText(string(txt))}
		structuredResult = &v
	case []any:
		resultArray = v
	case ToolResultMedia, ToolResultText:
		resultArray = []any{v}
	case []ToolResultMedia:
		resultArray = make([]any, len(v))
		for i, item := range v {
			resultArray[i] = item
		}
	case []ToolResultText:
		resultArray = make([]any, len(v))
		for i, item := range v {
			resultArray[i] = item
		}
	default:
		return fmt.Errorf("invalid tool result type: %T", v)
	}

	completeResult := map[string]any{
		"content": resultArray,
	}

	if structuredResult != nil {
		completeResult["structuredContent"] = structuredResult
	}

	return rw.WriteResult(completeResult)
}

func (rw *ResponseWriter) WriteResult(result map[string]any) error {
	response := Response[int]{
		Jsonrpc: JsonRPC,
		ID:      rw.id,
		Result:  result,
	}

	return json.NewEncoder(rw.w).Encode(response)
}

func (rw *ResponseWriter) WriteError(code int, message string) error {
	errorResponse := Response[int]{
		Jsonrpc: JsonRPC,
		ID:      rw.id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	return json.NewEncoder(rw.w).Encode(errorResponse)
}

func (rw *ResponseWriter) WriteToolError(message string) error {
	return rw.WriteResult(map[string]any{
		"content": []any{NewToolResultText(message)},
		"isError": true,
	})
}
