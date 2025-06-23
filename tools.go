package mcprobot

import (
	"encoding/base64"
	"fmt"
)

type ToolFunc[TR ToolResult] func(params map[string]any) (TR, error)

func (f ToolFunc[TR]) MCPHandler() MCPHandler {
	return MCPHandlerFunc(func(w RPCResponseWriter, req *Request[int]) {
		rw := NewResponseWriter(w, req.ID)

		errfmt := "failed to write response for reqID: %v: %s"
		args, ok := req.Params["arguments"].(map[string]any)
		if !ok {
			args = make(map[string]any)
		}

		result, err := f(args)
		if err != nil {
			switch e := err.(type) {
			case *ProtocolError:
				rw.WriteError(e.Code, e.Message)
			default:
				rw.WriteToolError(fmt.Sprintf(errfmt, req.ID, err))
			}
			return
		}

		if err := rw.WriteToolResult(result); err != nil {
			rw.WriteToolError(fmt.Sprintf(errfmt, req.ID, err))
			return
		}
	})
}

type ToolBuilder struct {
	definition ToolDefinition
}

func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		definition: ToolDefinition{
			Name:        name,
			InputSchema: make(map[string]any),
		},
	}
}

func (b *ToolBuilder) Description(description string) *ToolBuilder {
	b.definition.Description = description
	return b
}

func (b *ToolBuilder) initInputSchema() {
	if b.definition.InputSchema["properties"] == nil {
		b.definition.InputSchema["properties"] = make(map[string]any)
		b.definition.InputSchema["type"] = "object"
	}
}

func (b *ToolBuilder) withPropertyType(name, description, propertyType string, required bool) *ToolBuilder {
	b.initInputSchema()

	props := b.definition.InputSchema["properties"].(map[string]any)
	props[name] = map[string]any{
		"type":        propertyType,
		"description": description,
	}

	if required {
		if b.definition.Required == nil {
			b.definition.Required = []string{}
		}
		b.definition.Required = append(b.definition.Required, name)
	}

	return b
}

func (b *ToolBuilder) WithStringProperty(name, description string, required bool) *ToolBuilder {
	return b.withPropertyType(name, description, "string", required)
}

func (b *ToolBuilder) WithNumberProperty(name, description string, required bool) *ToolBuilder {
	return b.withPropertyType(name, description, "number", required)
}

func (b *ToolBuilder) WithBooleanProperty(name, description string, required bool) *ToolBuilder {
	return b.withPropertyType(name, description, "boolean", required)
}

func (b *ToolBuilder) WithArrayProperty(name, description string, required bool) *ToolBuilder {
	return b.withPropertyType(name, description, "array", required)
}

func (t *ToolBuilder) initAnnotations() {
	if t.definition.Annotations == nil {
		t.definition.Annotations = &ToolAnnotations{}
	}
}

func (b *ToolBuilder) Title(title string) *ToolBuilder {
	b.initAnnotations()
	b.definition.Annotations.Title = title
	return b
}

// IsReadOnly sets a hint that the tool is read-only,
// meaning it does not modify its environment
//
// Default value is not set
func (b *ToolBuilder) MarkReadOnly(readOnly bool) *ToolBuilder {
	b.initAnnotations()
	b.definition.Annotations.ReadOnlyHint = &readOnly
	return b
}

// MarkAsDistructive sets a hint that the tool may perform destructive updates
// (only meaningful when readOnlyHint is false)
//
// Default value is not set
func (b *ToolBuilder) MarkAsDistructive(isDestructive bool) *ToolBuilder {
	b.initAnnotations()
	b.definition.Annotations.DestructiveHint = &isDestructive
	return b
}

// MarkAsIdempotent sets a hint that calling the tool repeatedly
// with the same arguments has no additional effect
// (only meaningful when readOnlyHint is false)
//
// Default value is not set
func (b *ToolBuilder) MarkAsIdempotent(isIdempotent bool) *ToolBuilder {
	b.initAnnotations()
	b.definition.Annotations.IdempotentHint = &isIdempotent
	return b
}

// MarkAsCallingOpenWorld sets a hint that the tool interacts with
// a closed system (like a database) or an open system (like the web)
//
// Default value is not set
func (b *ToolBuilder) MarkAsCallingOpenWorld(callsOpenWorld bool) *ToolBuilder {
	b.initAnnotations()
	b.definition.Annotations.OpenWorldHint = &callsOpenWorld
	return b
}

func (b *ToolBuilder) Build() ToolDefinition {
	return b.definition
}

type (
	ToolDefinition struct {
		Name        string           `json:"name"`
		Description string           `json:"description"`
		InputSchema map[string]any   `json:"inputSchema"`
		Required    []string         `json:"required,omitempty"`
		Annotations *ToolAnnotations `json:"annotations,omitempty"`
	}

	ToolAnnotations struct {
		Title           string `json:"title,omitempty"`
		ReadOnlyHint    *bool  `json:"readOnlyHint,omitempty"`
		DestructiveHint *bool  `json:"destructiveHint,omitempty"`
		IdempotentHint  *bool  `json:"idempotentHint,omitempty"`
		OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`
	}
)

func (d ToolDefinition) validateArguments(args map[string]any) error {
	for _, field := range d.Required {
		if _, ok := args[field]; !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("missing required argument: %s", field),
			)
		}
	}

	schemaProps, ok := d.InputSchema["properties"].(map[string]any)
	if !ok {
		return nil
	}

	for argName, arg := range args {
		schema, exists := schemaProps[argName]
		if !exists {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("unexpected argument: %s", argName),
			)
		}

		schemaMap, ok := schema.(map[string]any)
		if !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("invalid schema for argument: %s. %T", argName, schema),
			)
		}

		expectedType, ok := schemaMap["type"].(string)
		if !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("invalid schema for argument: %s. %T", argName, schemaMap["type"]),
			)
		}

		if err := validateArgumentType(argName, arg, expectedType); err != nil {
			return err
		}

	}

	return nil
}

func validateArgumentType(name string, arg any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := arg.(string); !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be a string", name),
			)
		}
	case "number":
		if _, ok := arg.(float64); !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be a number", name),
			)
		}
	case "boolean":
		if _, ok := arg.(bool); !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be a boolean", name),
			)
		}
	case "array":
		if _, ok := arg.([]any); !ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be an array", name),
			)
		}
	default:
		if _, ok := arg.(map[string]any); ok {
			return NewProtocolError(
				ErrorCodeInvalidParams,
				fmt.Sprintf("object arguments are not supported. argument '%s' is an object. ", name),
			)
		}
	}

	return nil
}

type (
	ToolResult interface {
		ToolResultText |
			ToolResultMedia |
			[]ToolResultText |
			[]ToolResultMedia
	}

	ToolResultMedia struct {
		Type     string `json:"type"`
		Data     string `json:"data"`
		MimeType string `json:"mimeType"`
	}

	ToolResultText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	ToolResultStructured map[string]any
)

func NewToolResultText(text string) ToolResultText {
	return ToolResultText{
		Type: "text",
		Text: text,
	}
}

func NewToolResultImage(b []byte, mimeType string) ToolResultMedia {
	data := base64.StdEncoding.EncodeToString(b)

	return ToolResultMedia{
		Type:     "image",
		Data:     data,
		MimeType: mimeType,
	}
}

func NewToolResultAudio(b []byte, mimeType string) ToolResultMedia {
	data := base64.StdEncoding.EncodeToString(b)

	return ToolResultMedia{
		Type:     "audio",
		Data:     data,
		MimeType: mimeType,
	}
}
