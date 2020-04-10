package tests

import (
	"bytes"
	"testing"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	inputData = []byte{
		84,    // AInt
		2, 42, // OptInt (valid)
		0, // NilInt (nil)

		// hidden is omitted

		// Node :: Name
		24, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '-', 'n', 'o', 'd', 'e',

		// *Node :: Address (valid)
		2,

		2, // Id
		// *Next :: Address (valid)
		2,

		4, // Id
		// * Next :: Address (valid)
		2,

		6, // Id
		// *Next :: Address (nil)
		0,

		// OptAddr :: Address (valid)
		2,
		8, // Id
		0, // Next (nil)
	}

	unorderedSetterTestSchema = `{
		"name": "SetterTestRecord",
		"type": "record",
		"fields": [{
			"name": "aInt",
			"type": "int"
		}, {
			"name": "optInt",
			"type": ["null", "int"]
		}, {
			"name": "nilInt",
			"type": ["null", "int"]
		}, {
			"name": "node",
			"type": {
				"name": "Node",
				"type":	"record",
				"fields": [{
					"name": "name",
					"type": "string"
				}, {
					"name": "addr",
					"type": ["null", {
						"name": "address",
						"type": "record",
						"fields": [{
							"name": "id",
							"type": "int"
						}, {
							"name": "next",
							"type": ["null", "address"]
						}]
					}]
				}]
			}
		}, {
			"name": "optAddr",
			"type": ["null", "address"]
		}]
	}`

	reorderedSetterTestSchema = `{
		"name": "SetterTestRecord",
		"type": "record",
		"fields": [{
			"name": "optAddr",
			"type": ["null", "address"]
		}, {
			"name": "node",
			"type": {
				"name": "Node",
				"type":	"record",
				"fields": [{
					"name": "name",
					"type": "string"
				}, {
					"name": "addr",
					"type": ["null", {
						"name": "address",
						"type": "record",
						"fields": [{
							"name": "id",
							"type": "int"
						}, {
							"name": "next",
							"type": ["null", "address"]
						}]
					}]
				}]
			}
		}, {
			"name": "nilInt",
			"type": ["null", "int"]
		},  {
			"name": "optInt",
			"type": ["null", "int"]
		}, {
			"name": "aInt",
			"type": "int"
		}]
	}`
)

func TestUnorderedSetter(t *testing.T) {
	p, err := compiler.CompileSchemaBytes([]byte(unorderedSetterTestSchema), []byte(unorderedSetterTestSchema))
	require.Nil(t, err)

	var obj generated.OrderedSetterTestRecord
	executeSetterTest(t, p, &obj)

	assert.Equal(t, int32(42), obj.AInt)
	require.NotNil(t, obj.OptInt)
	assert.Equal(t, int32(21), *obj.OptInt)

	assert.Nil(t, obj.NilInt)

	assert.Equal(t, "example-node", obj.Node.Name)

	require.NotNil(t, obj.Node.Addr)
	assert.Equal(t, int32(1), obj.Node.Addr.Id)

	require.NotNil(t, obj.Node.Addr.Next)
	assert.Equal(t, int32(2), obj.Node.Addr.Next.Id)

	require.NotNil(t, obj.Node.Addr.Next.Next)
	assert.Equal(t, int32(3), obj.Node.Addr.Next.Next.Id)

	assert.Nil(t, obj.Node.Addr.Next.Next.Next)

	require.NotNil(t, obj.OptAddr)
	assert.Equal(t, int32(4), obj.OptAddr.Id)
	assert.Nil(t, obj.OptAddr.Next)

}

func TestReorderedSetter(t *testing.T) {
	p, err := compiler.CompileSchemaBytes([]byte(unorderedSetterTestSchema), []byte(reorderedSetterTestSchema))
	require.Nil(t, err)

	var obj generated.ReorderedSetterTestRecord
	executeSetterTest(t, p, &obj)

	assert.Equal(t, int32(42), obj.AInt)
	require.NotNil(t, obj.OptInt)
	assert.Equal(t, int32(21), *obj.OptInt)

	assert.Nil(t, obj.NilInt)

	assert.Equal(t, "example-node", obj.Node.Name)

	require.NotNil(t, obj.Node.Addr)
	assert.Equal(t, int32(1), obj.Node.Addr.Id)

	require.NotNil(t, obj.Node.Addr.Next)
	assert.Equal(t, int32(2), obj.Node.Addr.Next.Id)

	require.NotNil(t, obj.Node.Addr.Next.Next)
	assert.Equal(t, int32(3), obj.Node.Addr.Next.Next.Id)

	assert.Nil(t, obj.Node.Addr.Next.Next.Next)

	require.NotNil(t, obj.OptAddr)
	assert.Equal(t, int32(4), obj.OptAddr.Id)
	assert.Nil(t, obj.OptAddr.Next)
}

func executeSetterTest(t *testing.T, p vm.Program, objAddress interface{}) {
	engine := vm.Engine{
		Program:     p,
		StackTraces: true,
	}
	if err := engine.Run(bytes.NewBuffer(inputData), objAddress); err != nil {
		t.Fatalf("Program failed: %v", err)
	}
}
