package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnum(t *testing.T) {
	jsonString := `{
		"type": "enum",
		"name": "EnumTest",
		"symbols": ["TestSymbol1", "testSymbol2", "testSymbol3"],
		"aliases": ["e", "com.acme.Enum"]
	}`
	ns := NewNamespace(false)
	topLevel, err := ns.ParseSchema([]byte(jsonString))
	require.Nil(t, err)
	require.NotNil(t, topLevel)
}

func TestFixed(t *testing.T) {
	jsonString := `{
		"type": "fixed",
		"name": "FixedTest",
		"size": 16,
		"aliases": ["f", "com.acme.Fixed"]
	}`
	ns := NewNamespace(false)
	topLevel, err := ns.ParseSchema([]byte(jsonString))
	require.Nil(t, err)
	require.NotNil(t, topLevel)
}

func TestMap(t *testing.T) {
	jsonString := `{
		"type": "map",
		"values": {
			"type": "enum",
			"name": "MapEnumTest",
			"symbols": ["TestSymbol1", "testSymbol2", "testSymbol3"],
			"aliases": ["e", "com.acme.Enum"]
		}
	}`
	ns := NewNamespace(false)
	topLevel, err := ns.ParseSchema([]byte(jsonString))
	require.Nil(t, err)
	require.NotNil(t, topLevel)
}

func TestArray(t *testing.T) {
	jsonString := `{
		"type": "array",
		"items": {
			"type": "enum",
			"name": "ArrayEnumTest",
			"symbols": ["TestSymbol1", "testSymbol2", "testSymbol3"],
			"aliases": ["e", "com.acme.Enum"]
		}
	}`
	ns := NewNamespace(false)
	topLevel, err := ns.ParseSchema([]byte(jsonString))
	require.Nil(t, err)
	require.NotNil(t, topLevel)
}

func TestRecord(t *testing.T) {
	jsonString := `{
		"type" : "record",
		"name" : "AliasRecord",
		"fields" : [ 
		{
			"name": "a",
			"type": {
				"type": "enum",
				"name": "ArrayEnumTest",
				"symbols": ["TestSymbol1", "testSymbol2", "testSymbol3"],
				"aliases": ["e", "com.acme.Enum"]
			}
		},
		{
			"name": "c",
			"aliases": ["d"],
			"type": {
				"type": "array",
				"items": "com.acme.Enum"
			}
		}
		]
	}`
	ns := NewNamespace(false)
	topLevel, err := ns.ParseSchema([]byte(jsonString))
	require.Nil(t, err)
	require.NotNil(t, topLevel)
}

func TestAliasedRecord(t *testing.T) {
	jsonString := `{
	"type" : "record",
	"name" : "AliasedRecord",
	"fields" : [{
		"type": "record",
		"name": "NestedRecord",
		"fields": [{
			"name": "OtherField",
			"type": "aliasedRecord"
		}]
	},{
		"type": "record",
		"name": "MasterRecord",
		"aliases": [
			"aliasedRecord"
		],
		"fields": [{
			"name": "StringField",
			"type": "string"
		}, {
			"name": "BoolField",
			"type": "boolean"
		}, {
			"name": "BytesField",
			"type": "bytes"
		}]
	}]
}`
	ns := NewNamespace(false)
	topLevel, err := ns.ParseSchema([]byte(jsonString))
	require.Nil(t, err)
	require.NotNil(t, topLevel)
}
