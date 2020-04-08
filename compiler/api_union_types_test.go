package compiler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrimitiveUnionTypes(t *testing.T) {
	schemas := []string{
		`["boolean", "int", "long", "float", "double", "string", "bytes"]`,
	}

	unionsRoundtrip(t, "union_primitives-%d.asm", schemas, schemas)
}

func TestEquivalentUnionPrimitiveTypes(t *testing.T) {
	writerSchemas := []string{
		`["boolean", "int"]`,
		`["boolean", "int"]`,
		`["boolean", "int"]`,
		`["boolean", "int"]`,

		`["boolean", "long"]`,
		`["boolean", "long"]`,
		`["boolean", "long"]`,

		`["boolean", "float"]`,
		`["boolean", "float"]`,

		`["boolean", "double"]`,

		`["long", "string"]`,
		`["long", "string"]`,
	}

	readerSchemas := []string{
		`["int", "boolean"]`,
		`["long", "boolean"]`,
		`["float", "boolean"]`,
		`["double", "boolean"]`,

		`["long", "boolean"]`,
		`["float", "boolean"]`,
		`["double", "boolean"]`,

		`["float", "boolean"]`,
		`["double", "boolean"]`,

		`["double", "boolean"]`,

		`["string", "long"]`,
		`["bytes", "long"]`,
	}

	unionsRoundtrip(t, "union_equiv_%d.asm", writerSchemas, readerSchemas)
}

func TestOptionalUnionTypes(t *testing.T) {
	schemas := []string{
		// Optional-first, single-typed unions
		`["null", "boolean"]`,
		`["null", "int"]`,
		`["null", "long"]`,
		`["null", "float"]`,
		`["null", "double"]`,
		`["null", "string"]`,
		`["null", "bytes"]`,

		// Optional-nonfirst, single-typed unions
		`["boolean", "null"]`,
		`["int", "null"]`,
		`["long", "null"]`,
		`["float", "null"]`,
		`["double", "null"]`,
		`["string", "null"]`,
		`["bytes", "null"]`,

		// Optional-first, multi-typed unions
		`["null", "boolean", "int", "long"]`,
		`["null", "float", "double", "string", "bytes"]`,

		// Optional-nonfirst, multi-typed unions
		`["boolean", "null", "int", "long"]`,
		`["float", "null", "double", "string", "bytes"]`,
	}

	unionsRoundtrip(t, "union_opt-%d.asm", schemas, schemas)
}

func TestUnionToNonUnionPrimitiveTypes(t *testing.T) {
	writerSchemas := []string{
		`["boolean", "int"]`,
		`["boolean", "int"]`,
		`["boolean", "int"]`,
		`["boolean", "int"]`,

		`["boolean", "long"]`,
		`["boolean", "long"]`,
		`["boolean", "long"]`,

		`["boolean", "float"]`,
		`["boolean", "float"]`,

		`["boolean", "double"]`,

		`["long", "string"]`,
		`["long", "string"]`,

		`["long", "bytes"]`,
		`["long", "bytes"]`,
	}

	readerSchemas := []string{
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aInt", "type": "int"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aLong", "type": "long"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aFloat", "type": "float"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aDouble", "type": "double"}, { "name": "aBool", "type": "boolean"} ] }`,

		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aLong", "type": "long"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aFloat", "type": "float"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aDouble", "type": "double"}, { "name": "aBool", "type": "boolean"} ] }`,

		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aFloat", "type": "float"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aDouble", "type": "double"}, { "name": "aBool", "type": "boolean"} ] }`,

		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aDouble", "type": "double"}, { "name": "aBool", "type": "boolean"} ] }`,

		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aString", "type": "string"}, { "name": "aLong", "type": "long"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aBytes", "type": "bytes"}, { "name": "aLong", "type": "long"} ] }`,

		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aString", "type": "string"}, { "name": "aLong", "type": "long"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aBytes", "type": "bytes"}, { "name": "aLong", "type": "long"} ] }`,
	}

	unionsRoundtrip(t, "union_to_non-union_%d.asm", writerSchemas, readerSchemas)
}

func unionsRoundtrip(t *testing.T, nameFormat string, writerSchemas, readerSchemas []string) {
	for i := range writerSchemas {
		prog, err := CompileSchemaBytes([]byte(writerSchemas[i]), []byte(readerSchemas[i]))
		require.Nil(t, err)

		goldenEquals(t, fmt.Sprintf(nameFormat, i), []byte(prog.String()))
	}
}
