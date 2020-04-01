package schema

import "encoding/json"

type SchemaType interface {
	Schema() string
	SetSchema(schemaMap map[string]interface{})
}

// Common attributes for all schema-holding types
type schemaComponent struct {
	schema string
}

var (
	// Ensure interface implementation
	_ SchemaType = &schemaComponent{}
)

func (comp *schemaComponent) Schema() string {
	return comp.schema
}

func (comp *schemaComponent) SetSchema(schemaMap map[string]interface{}) {
	jsonBytes, _ := json.Marshal(schemaMap)
	comp.schema = string(jsonBytes)
}
