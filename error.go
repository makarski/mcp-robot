package mcprobot

const (
	ErrorCodeInvalidRequest      = -32600
	ErrorCodeMethodNotFound      = -32601
	ErrorCodeInvalidParams       = -32602
	ErrorCodeInternalError       = -32603
	ErrorCodeToolNotFound        = -32000
	ErrorCodeToolExecutionFailed = -32001
	ErrorCodeParseError          = -32700
)

type ProtocolError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewProtocolError(code int, message string) *ProtocolError {
	return &ProtocolError{
		Code:    code,
		Message: message,
	}
}

func (e *ProtocolError) Error() string {
	return e.Message
}
