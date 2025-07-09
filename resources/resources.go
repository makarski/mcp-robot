package resources

type ResourceType string

const (
	ResourceItemType ResourceType = "resource"
	ResourceLinkType ResourceType = "resource_link"
)

type (
	Resource interface {
		ResourceText | ResourceBinary
	}

	resource struct {
		URI         string `json:"uri"`
		Name        string `json:"name"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		MimeType    string `json:"mimeType,omitempty"`
		Size        int64  `json:"size,omitempty"`
	}

	ResourceText struct {
		resource
		Text string `json:"text"`
	}

	ResourceBinary struct {
		resource
		// Base64 encoded binary data
		Blob []byte `json:"blob"`
	}

	// Only used in Tools for now
	ResourceLink struct {
		Type ResourceType `json:"type"`
		resource
	}
)

func NewResourceText(uri, name, title, description, mimeType, text string) ResourceText {
	return ResourceText{
		resource: resource{
			URI:         uri,
			Name:        name,
			Title:       title,
			Description: description,
			MimeType:    mimeType,
		},
		Text: text,
	}
}

func NewResourceBinary(uri, name, title, description, mimeType string, blob []byte) ResourceBinary {
	return ResourceBinary{
		resource: resource{
			URI:         uri,
			Name:        name,
			Title:       title,
			Description: description,
			MimeType:    mimeType,
			Size:        int64(len(blob)),
		},
		Blob: blob,
	}
}

func NewResourceLink(uri, name, description, mimeType string) ResourceLink {
	return ResourceLink{
		Type: ResourceLinkType,
		resource: resource{
			URI:         uri,
			Name:        name,
			Description: description,
			MimeType:    mimeType,
		},
	}
}
