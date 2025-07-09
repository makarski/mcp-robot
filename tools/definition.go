package tools

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/makarski/mcp-robot/spec"
)

type (
	ToolDefinition struct {
		Name         string          `json:"name"`
		Title        string          `json:"title,omitempty"`
		Description  string          `json:"description"`
		InputSchema  ToolSchema      `json:"inputSchema,omitzero"`
		OutputSchema ToolSchema      `json:"outputSchema,omitzero"`
		Annotations  ToolAnnotations `json:"annotations,omitzero"`

		ResultsPerPage int `json:"-"`
	}

	ToolSchema struct {
		Type        string                 `json:"type"`
		Description string                 `json:"description,omitempty"`
		Properties  map[string]*ToolSchema `json:"properties,omitempty"`
		Required    []string               `json:"required,omitempty"`
		ArrayItems  *ToolSchema            `json:"items,omitempty"`
	}

	ToolAnnotations struct {
		Title           string `json:"title,omitempty"`
		ReadOnlyHint    *bool  `json:"readOnlyHint,omitempty"`
		DestructiveHint *bool  `json:"destructiveHint,omitempty"`
		IdempotentHint  *bool  `json:"idempotentHint,omitempty"`
		OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`
	}
)

func (d ToolDefinition) ValidateArguments(args map[string]any) error {
	for _, field := range d.InputSchema.Required {
		if _, ok := args[field]; !ok {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("missing required argument: %s", field),
			)
		}
	}

	if d.InputSchema.Properties == nil {
		return nil
	}

	for argName, arg := range args {
		schema, exists := d.InputSchema.Properties[argName]
		if !exists {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("unexpected argument: %s", argName),
			)
		}

		if err := validateArgumentType(argName, arg, schema.Type); err != nil {
			return err
		}

	}

	return nil
}

func validateArgumentType(name string, arg any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := arg.(string); !ok {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be a string", name),
			)
		}
	case "number":
		if _, ok := arg.(float64); !ok {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be a number", name),
			)
		}
	case "boolean":
		if _, ok := arg.(bool); !ok {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be a boolean", name),
			)
		}
	case "array":
		v := reflect.ValueOf(arg)
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("argument '%s' must be an array, got %T", name, arg),
			)
		}
	default:
		if _, ok := arg.(map[string]any); ok {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("object arguments are not supported. argument '%s' is an object. ", name),
			)
		}
	}

	return nil
}

func validateOutput[TR ToolResult](definition ToolDefinition, result TR) error {
	schema := definition.OutputSchema
	if schema.Type == "" {
		return nil // No output schema defined, no validation needed
	}

	switch v := any(result).(type) {
	case ToolResultStructured:
		if len(schema.Required) == 0 && len(schema.Properties) == 0 {
			return nil
		}
		missingFields := []string{}
		for _, requiredField := range schema.Required {
			if _, ok := v[requiredField]; !ok {
				missingFields = append(missingFields, requiredField)
			}
		}

		if len(missingFields) > 0 {
			return spec.NewProtocolError(
				spec.ErrorCodeInvalidParams,
				fmt.Sprintf("missing required output fields: %v", strings.Join(missingFields, ",")),
			)
		}

		for name, spec := range schema.Properties {
			if _, ok := v[name]; !ok {
				continue
			}

			if err := validateArgumentType(name, v[name], spec.Type); err != nil {
				return err
			}

			// TODO: Handle nested objects and arrays
			// if spec.Type == "array" {
			// arrayVals := reflect.ValueOf(v[name])
			// if arrayVals.Kind() == reflect.Slice && arrayVals.Kind() == reflect.Array {
			// 	for i := 0; i < arrayVals.Len(); i++ {
			// 		itemVal := arrayVals.Index(i).Interface()

			// for _, arrayItem := range spec.arrayItems.Properties {
			// fmt.Printf("array item: %+v, %+v", arrayItem, v[name])
			// }
			// }
		}
	}

	return nil
}
