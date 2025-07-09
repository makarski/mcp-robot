package tools

type (
	ToolBuilder struct {
		definition ToolDefinition
	}

	SchemaBuilder struct {
		root        *ToolBuilder
		parent      *SchemaBuilder
		schemaField *ToolSchema
	}

	ArrayBuilder struct {
		parent      *SchemaBuilder
		name        string
		description string
		schemaField *ToolSchema
	}

	ObjectBuilder struct {
		parent      *SchemaBuilder
		name        string
		description string
		schemaField *ToolSchema
	}
)

func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		definition: ToolDefinition{
			Name: name,
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: make(map[string]*ToolSchema),
				Required:   make([]string, 0),
			},
		},
	}
}

func (b *ToolBuilder) Input() *SchemaBuilder {
	sb := &SchemaBuilder{root: b, schemaField: &b.definition.InputSchema}
	sb.initSchema()
	return sb
}

func (b *ToolBuilder) Output() *SchemaBuilder {
	sb := &SchemaBuilder{root: b, schemaField: &b.definition.OutputSchema}
	sb.initSchema()
	return sb
}

func (b *ToolBuilder) Description(description string) *ToolBuilder {
	b.definition.Description = description
	return b
}

func (b *SchemaBuilder) initSchema() {
	if b.schemaField.Properties == nil {
		b.schemaField.Properties = make(map[string]*ToolSchema)
		b.schemaField.Type = "object"
		b.schemaField.Required = make([]string, 0)
	}
}

func (sb *SchemaBuilder) withPropertyType(name, description, propertyType string, required bool) *SchemaBuilder {
	sb.schemaField.Properties[name] = &ToolSchema{
		Type:        propertyType,
		Description: description,
	}

	if required {
		sb.schemaField.Required = append(sb.schemaField.Required, name)
	}

	return sb
}

func (sb *SchemaBuilder) WithString(name, description string, required bool) *SchemaBuilder {
	return sb.withPropertyType(name, description, "string", required)
}

func (sb *SchemaBuilder) WithNumber(name, description string, required bool) *SchemaBuilder {
	return sb.withPropertyType(name, description, "number", required)
}

func (sb *SchemaBuilder) WithBoolean(name, description string, required bool) *SchemaBuilder {
	return sb.withPropertyType(name, description, "boolean", required)
}

func (sb *SchemaBuilder) WithArray(name, description string, required bool) *ArrayBuilder {
	ab := &ArrayBuilder{
		parent:      sb,
		name:        name,
		description: description,
		schemaField: nil,
	}

	if required {
		sb.schemaField.Required = append(sb.schemaField.Required, name)
	}

	return ab
}

func (ab *ArrayBuilder) Of(propertyType, description string) *SchemaBuilder {
	innerArrayItem := ToolSchema{
		Type:        propertyType,
		Description: description,
		Properties:  make(map[string]*ToolSchema),
		Required:    make([]string, 0),
	}

	itemsNode := ToolSchema{
		Type:        "array",
		Description: ab.description,
		Required:    make([]string, 0),
		ArrayItems:  &innerArrayItem,
	}

	ab.parent.schemaField.Properties[ab.name] = &itemsNode
	nestedArrayBuilder := &SchemaBuilder{
		root:        ab.parent.root,
		parent:      ab.parent,
		schemaField: &innerArrayItem,
	}

	return nestedArrayBuilder
}

func (sb *SchemaBuilder) WithObject(name, description string, required bool) *ObjectBuilder {
	ob := &ObjectBuilder{
		parent:      sb,
		name:        name,
		description: description,
		schemaField: nil,
	}

	if required {
		sb.schemaField.Required = append(sb.schemaField.Required, name)
	}

	return ob
}

func (ob *ObjectBuilder) Props() *SchemaBuilder {
	nestedSchemaField := ToolSchema{
		Type:        "object",
		Description: ob.description,
		Properties:  make(map[string]*ToolSchema),
		Required:    make([]string, 0),
	}

	ob.parent.schemaField.Properties[ob.name] = &nestedSchemaField

	nestedBuilder := &SchemaBuilder{root: ob.parent.root, parent: ob.parent, schemaField: &nestedSchemaField}
	return nestedBuilder
}

func (sb *SchemaBuilder) Done() *ToolBuilder {
	return sb.root
}

func (b *ToolBuilder) Title(title string) *ToolBuilder {
	b.definition.Annotations.Title = title
	return b
}

// IsReadOnly sets a hint that the tool is read-only,
// meaning it does not modify its environment
//
// Default value is not set
func (b *ToolBuilder) MarkReadOnly(readOnly bool) *ToolBuilder {
	b.definition.Annotations.ReadOnlyHint = &readOnly
	return b
}

// MarkAsDestructive sets a hint that the tool may perform destructive updates
// (only meaningful when readOnlyHint is false)
//
// Default value is not set
func (b *ToolBuilder) MarkAsDestructive(isDestructive bool) *ToolBuilder {
	b.definition.Annotations.DestructiveHint = &isDestructive
	return b
}

// MarkAsIdempotent sets a hint that calling the tool repeatedly
// with the same arguments has no additional effect
// (only meaningful when readOnlyHint is false)
//
// Default value is not set
func (b *ToolBuilder) MarkAsIdempotent(isIdempotent bool) *ToolBuilder {
	b.definition.Annotations.IdempotentHint = &isIdempotent
	return b
}

// MarkAsCallingOpenWorld sets a hint that the tool interacts with
// a closed system (like a database) or an open system (like the web)
//
// Default value is not set
func (b *ToolBuilder) MarkAsCallingOpenWorld(callsOpenWorld bool) *ToolBuilder {
	b.definition.Annotations.OpenWorldHint = &callsOpenWorld
	return b
}

func (b *ToolBuilder) Build() ToolDefinition {
	return b.definition
}
