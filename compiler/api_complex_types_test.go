package compiler

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileComplex(t *testing.T) {
	wrtSchema := `{
			"type": "record",
			"name": "TestRecord",
			"fields": [ {
				"name": "aBytes",
				"type": "bytes"
			}, {
				"name": "aRecord",
				"type": {
					"name": "SimpleRecord",
					"type": "record",
					"fields": [
						{
							"name": "innerInt",
							"type": "int"
						}
					]
				}
			}, {
				"name": "aString",
				"type": "string"
			}, {
				"name": "aFloat",
				"type": "float"
			}, {
				"name": "aInt",
				"type": "int"
			}, {
				"name": "aIntMap",
				"type": {"type": "map", "values": "int"}
			} ]
		}`
	wrtData := []byte{
		8, 0, 1, 2, 3, // bytes: {0, 1, 2, 3}
		84,               // SimpleRecord::innerInt: 42
		6, 'H', 'i', '!', // string: 'Hi!'
		0, 1, 2, 3, // float
		42, // int: 21
		4,
		6, 'O', 'n', 'e',
		2,
		6, 'T', 'w', 'o',
		4,
		0,
	}

	rdrSchemas := []string{
		/*
			`{
				"type": "record",
				"name": "TestRecord",
				"fields": [ {
					"name": "aBool",
					"type": "bool"
				}, {
					"name": "aRecord",
					"type": {
						"name": "SimpleRecord",
						"type": "record",
						"fields": [
							{
								"name": "innerInt",
								"type": "int"
							}
						]
					}
				}, {
					"name": "aString",
					"type": "string"
				}, {
					"name": "aEnum",
					"type": "enum",
					"values": ["A", "B", "C"]
				}, {
					"name": "aInt",
					"type": "int"
				}, {
					"name": "aIntMap",
					"type": {"type": "map", "values": "int"}
				} ]
			}`,
			`{
				"type": "record",
				"name": "TestRecord",
				"fields": [ {
					"name": "aInt",
					"type": "int"
				}, {
					"name": "aBool",
					"type": "bool"
				}, {
					"name": "aRecord",
					"type": {
						"name": "SimpleRecord",
						"type": "record",
						"fields": [
							{
								"name": "innerInt",
								"type": "int"
							}
						]
					}
				}, {
					"name": "aIntMap",
					"type": {"type": "map", "values": "int"}
				}, {
					"name": "aString",
					"type": "string"
				}, {
					"name": "aEnum",
					"type": "enum",
					"values": ["A", "B", "C"]
				} ]
			}`,*/
		`{
			"type": "record",
			"name": "TestRecord",
			"fields": [ {
				"name": "aInt",
				"type": "int"
			} ]
		}`,
	}

	for i := range rdrSchemas {
		p, err := CompileSchemaBytes([]byte(wrtSchema), []byte(rdrSchemas[i]))
		require.Nil(t, err)
		fmt.Printf("Program:\n%s\n", p)

		var obj struct {
			Placeholder int32
		}
		objSetter, err := setters.NewSetterFor(&obj)
		if err != nil {
			t.Fatal(err)
		}

		engine := vm.NewEngine(p, objSetter)
		err = engine.Run(bytes.NewBuffer(wrtData))
		require.Nil(t, err)
		assert.Equal(t, int32(21), obj.Placeholder)
	}
}
