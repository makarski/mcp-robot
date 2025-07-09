package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/makarski/mcp-robot/io"
	"github.com/makarski/mcp-robot/resources"
)

type (
	ToolResult interface {
		ToolResultText |
			ToolResultMedia |
			[]ToolResultText |
			[]ToolResultMedia |
			ToolResultStructured |
			resources.ResourceLink |
			ToolResultEmbeddedResource[resources.ResourceText] |
			ToolResultEmbeddedResource[resources.ResourceBinary] |
			[]ToolResultEmbeddedResource[resources.ResourceText] |
			[]ToolResultEmbeddedResource[resources.ResourceBinary] |
			ToolResultUnion
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

	ToolResultEmbeddedResource[R resources.Resource] struct {
		Type     resources.ResourceType `json:"type"`
		Resource R                      `json:"resource"`
	}

	ToolResultUnion struct {
		items []toolResultUnionItem
	}

	toolResultUnionItem struct {
		text                   *ToolResultText
		media                  *ToolResultMedia
		embeddedTextResource   *ToolResultEmbeddedResource[resources.ResourceText]
		embeddedBinaryResource *ToolResultEmbeddedResource[resources.ResourceBinary]
		resourceLink           *resources.ResourceLink
	}
)

func NewToolResultUnion() *ToolResultUnion {
	return &ToolResultUnion{
		items: make([]toolResultUnionItem, 0),
	}
}

func (u *ToolResultUnion) AddText(text string) *ToolResultUnion {
	txt := NewToolResultText(text)
	u.items = append(u.items, toolResultUnionItem{
		text: &txt,
	})
	return u
}

func (u *ToolResultUnion) AddImage(b []byte, mimeType string) *ToolResultUnion {
	image := NewToolResultImage(b, mimeType)
	item := toolResultUnionItem{
		media: &image,
	}
	u.items = append(u.items, item)
	return u
}

func (u *ToolResultUnion) AddAudio(b []byte, mimeType string) *ToolResultUnion {
	audio := NewToolResultAudio(b, mimeType)
	item := toolResultUnionItem{
		media: &audio,
	}
	u.items = append(u.items, item)
	return u
}

func (u *ToolResultUnion) AddEmbeddedTextResource(uri, name, description, mimeType, text string) *ToolResultUnion {
	resource := NewToolResultEmbeddedTextResource(uri, name, description, mimeType, text)
	item := toolResultUnionItem{
		embeddedTextResource: &resource,
	}
	u.items = append(u.items, item)
	return u
}

func (u *ToolResultUnion) AddEmbeddedBinaryResource(uri, name, description, mimeType string, b []byte) *ToolResultUnion {
	resource := NewToolResultEmbeddedBinaryResource(uri, name, description, mimeType, b)
	item := toolResultUnionItem{
		embeddedBinaryResource: &resource,
	}
	u.items = append(u.items, item)
	return u
}

func (u *ToolResultUnion) AddResourceLink(uri, name, description, mimeType string) *ToolResultUnion {
	link := resources.NewResourceLink(uri, name, description, mimeType)
	item := toolResultUnionItem{
		resourceLink: &link,
	}
	u.items = append(u.items, item)
	return u
}

func (u *ToolResultUnion) toArray() []any {
	result := make([]any, len(u.items))

	for i, item := range u.items {
		switch {
		case item.text != nil:
			result[i] = *item.text
		case item.media != nil:
			result[i] = *item.media
		case item.embeddedTextResource != nil:
			result[i] = *item.embeddedTextResource
		case item.embeddedBinaryResource != nil:
			result[i] = *item.embeddedBinaryResource
		case item.resourceLink != nil:
			result[i] = *item.resourceLink
		}
	}

	return result
}

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

func NewToolResultResourceLink(
	uri string,
	name string,
	description string,
	mimeType string,
) resources.ResourceLink {
	return resources.NewResourceLink(uri, name, description, mimeType)
}

func NewToolResultEmbeddedTextResource(
	uri string,
	name string,
	description string,
	mimeType string,
	text string,
) ToolResultEmbeddedResource[resources.ResourceText] {
	return ToolResultEmbeddedResource[resources.ResourceText]{
		Type:     resources.ResourceItemType,
		Resource: resources.NewResourceText(uri, name, "", description, mimeType, text),
	}
}

func NewToolResultEmbeddedBinaryResource(
	uri string,
	name string,
	description string,
	mimeType string,
	b []byte,
) ToolResultEmbeddedResource[resources.ResourceBinary] {
	return ToolResultEmbeddedResource[resources.ResourceBinary]{
		Type:     resources.ResourceItemType,
		Resource: resources.NewResourceBinary(uri, name, "", description, mimeType, b),
	}
}

func writeResult(rw *io.ResponseWriter, result any) error {
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
	case ToolResultMedia,
		ToolResultText,
		resources.ResourceLink,
		ToolResultEmbeddedResource[resources.ResourceBinary],
		ToolResultEmbeddedResource[resources.ResourceText]:
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
	case ToolResultUnion:
		resultArray = v.toArray()
	default:
		return fmt.Errorf("invalid tool result type: %T", v)
	}

	completeResult := map[string]any{
		"content": resultArray,
		"isError": false,
	}

	if structuredResult != nil {
		completeResult["structuredContent"] = structuredResult
	}

	return rw.WriteResult(completeResult)
}

func writeError(rw *io.ResponseWriter, message string) error {
	return rw.WriteResult(map[string]any{
		"content": []any{NewToolResultText(message)},
		"isError": true,
	})
}
