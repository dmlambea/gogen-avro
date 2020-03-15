package schema_test

import (
	"testing"

	"github.com/actgardner/gogen-avro/parser"

	"github.com/stretchr/testify/assert"
)

func TestIsReadableBy(t *testing.T) {
	cases := []struct {
		writer     string
		reader     string
		isReadable bool
	}{
		// Numeric types can be promoted to a larger size
		{`"int"`, `"int"`, true},
		{`"int"`, `"long"`, true},
		{`"int"`, `"float"`, true},
		{`"int"`, `"double"`, true},
		{`"long"`, `"long"`, true},
		{`"long"`, `"float"`, true},
		{`"long"`, `"double"`, true},
		{`"float"`, `"float"`, true},
		{`"float"`, `"double"`, true},
		{`"double"`, `"double"`, true},

		// Numeric types can't be demoted to a smaller size
		{`"long"`, `"int"`, false},
		{`"float"`, `"int"`, false},
		{`"float"`, `"long"`, false},
		{`"double"`, `"int"`, false},
		{`"double"`, `"long"`, false},
		{`"double"`, `"float"`, false},

		// String and bytes fields are interchangable
		{`"string"`, `"bytes"`, true},
		{`"bytes"`, `"string"`, true},

		// Record fields are matched by name
		{`{"type": "record", "name": "test", "fields": [{"name": "a", "type": "int"}]}`, `{"type": "record", "name": "test", "fields": [{"name": "a", "type": "long"}]}`, true},
		{`{"type": "record", "name": "test", "fields": [{"name": "a", "type": "int"}]}`, `{"type": "record", "name": "test", "fields": [{"name": "a", "type": "string"}]}`, false},

		// Any type can be promoted to a union of that type and another
		{`"boolean"`, `["boolean", "string"]`, true},
		{`"int"`, `["int", "string"]`, true},
		{`"long"`, `["long", "string"]`, true},
		{`"float"`, `["float", "string"]`, true},
		{`"double"`, `["double", "string"]`, true},
		{`"string"`, `["double", "string"]`, true},
		{`"bytes"`, `["double", "string"]`, true},
		{`{"type": "array", "items": "int"}`, `["string", {"type": "array", "items": "int"}]`, true},
		{`{"type": "map", "values": "int"}`, `["string", {"type": "map", "values": "int"}]`, true},
		{`{"type": "record", "name": "test", "fields": [{"name": "a", "type": "int"}]}`, `[{"type": "record", "name": "test", "fields": [{"name": "a", "type": "int"}]}, "string"]`, true},

		// A union can be read with a single type from that union, provided the reader is the "chosen" type
		{`["double", "string"]`, `"bytes"`, true},

		// An optional union can be read by another optional union
		{`["null", "string"]`, `["int", "null"]`, true},
	}

	for i, c := range cases {
		ns1 := parser.NewNamespace(false)
		writer, err := ns1.ParseSchema([]byte(c.writer))
		assert.Nil(t, err)

		ns2 := parser.NewNamespace(false)
		reader, err := ns2.ParseSchema([]byte(c.reader))
		assert.Nil(t, err)

		assert.Equal(t, c.isReadable, writer.IsReadableBy(reader, make(map[string]bool)), "Bug %d:\n  Writer: %s\n  Reader: %s", i+1, c.writer, c.reader)
	}
}
