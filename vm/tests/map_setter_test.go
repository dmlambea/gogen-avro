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

func TestSimpleMapSetter(t *testing.T) {
	simpleMapInputData := []byte{
		84,

		4, // Block length: 2 items

		8, 'k', 'e', 'y', '1',
		8, 'v', 'a', 'l', '1',

		8, 'k', 'e', 'y', '2',
		8, 'v', 'a', 'l', '2',

		0, // End of blocks
	}

	simpleMapSchema := `{
		"name": "SimpleMapRecord",
		"type": "record",
		"fields": [{
			"name": "aInt",
			"type": "int"
		}, {
			"name": "nodes",
			"type": {
				"type": "map",
				"values": {
					"name": "nested",
					"type":	"record",
					"fields": [{
						"name": "val",
						"type": "string"
					}]
				}
			}
		}]
	}`
	p, err := compiler.CompileSchemaBytes([]byte(simpleMapSchema), []byte(simpleMapSchema))
	require.Nil(t, err)

	engine := vm.Engine{
		Program:     p,
		StackTraces: true,
	}

	var obj generated.SimpleMapRecord
	if err = engine.Run(bytes.NewBuffer(simpleMapInputData), &obj); err != nil {
		t.Fatalf("Program failed: %v", err)
	}

	assert.Equal(t, int32(42), obj.AInt)

	require.Equal(t, 2, len(obj.Nodes))

	var value generated.NestedValue
	var ok bool

	value, ok = obj.Nodes["key1"]
	assert.True(t, ok)
	assert.Equal(t, "val1", value.Val)

	value, ok = obj.Nodes["key2"]
	assert.True(t, ok)
	assert.Equal(t, "val2", value.Val)
}

func TestComplexMapSetter(t *testing.T) {
	complexMapInputData := []byte{
		84,    // aInt: 42
		2, 42, // *optInt: 21

		4, // Block length: 2 items

		10, 'd', 'e', 'm', 'o', '1',

		00, 00, 12, 34,
		2, // Block length: 1 item
		14, 'o', 'n', 'e', '-', 'o', 'n', 'e', 22,
		0, // End of blocks

		10, 'd', 'e', 'm', 'o', '2',

		00, 00, 23, 45,
		6, // Block length: 3 item
		14, 't', 'w', 'o', '-', 'o', 'n', 'e', 42,
		14, 't', 'w', 'o', '-', 't', 'w', 'o', 44,
		18, 't', 'w', 'o', '-', 't', 'h', 'r', 'e', 'e', 46,
		0, // End of blocks

		0, // End of blocks
	}

	complexMapSchema := `{
		"name": "NestedMapRecord",
		"type": "record",
		"fields": [{
			"name": "aInt",
			"type": "int"
		}, {
			"name": "optInt",
			"type": ["null", "int"]
		}, {
			"name": "nodes",
			"type": {
				"type": "map",
				"values": {
					"name": "numberMap",
					"type":	"record",
					"fields": [{
						"name": "index",
						"type": "float"
					}, {
						"name": "numbers",
						"type": {
							"type": "map",
							"values": "int"
						}
					}]
				}
			}
		}]
	}`
	p, err := compiler.CompileSchemaBytes([]byte(complexMapSchema), []byte(complexMapSchema))
	require.Nil(t, err)

	engine := vm.Engine{
		Program:     p,
		StackTraces: true,
	}

	var obj generated.NestedMapRecord
	if err = engine.Run(bytes.NewBuffer(complexMapInputData), &obj); err != nil {
		t.Fatalf("Program failed: %v", err)
	}

	assert.Equal(t, int32(42), obj.AInt)

	require.NotNil(t, obj.OptInt)
	assert.Equal(t, int32(21), *obj.OptInt)

	require.Equal(t, 2, len(obj.Nodes))

	var r generated.NumberMap
	var ok bool
	r, ok = obj.Nodes["demo1"]
	assert.True(t, ok)
	assert.True(t, r.Index != 0)
	require.Equal(t, 1, len(r.Numbers))
	assert.Equal(t, int32(11), r.Numbers["one-one"])

	r, ok = obj.Nodes["demo2"]
	assert.True(t, ok)
	assert.True(t, r.Index != 0)
	require.Equal(t, 3, len(r.Numbers))
	assert.Equal(t, int32(21), r.Numbers["two-one"])
	assert.Equal(t, int32(22), r.Numbers["two-two"])
	assert.Equal(t, int32(23), r.Numbers["two-three"])
}

var benchMapErr error

func BenchmarkMapSetter(b *testing.B) {
	benchmarkMapInputData := []byte{
		84,    // aInt: 42
		2, 42, // *optInt: 21

		4, // Block length: 2 items

		10, 'd', 'e', 'm', 'o', '1',

		00, 00, 12, 34,
		2, // Block length: 1 item
		14, 'o', 'n', 'e', '-', 'o', 'n', 'e', 22,
		0, // End of blocks

		10, 'd', 'e', 'm', 'o', '2',

		00, 00, 23, 45,
		6, // Block length: 3 item
		14, 't', 'w', 'o', '-', 'o', 'n', 'e', 42,
		14, 't', 'w', 'o', '-', 't', 'w', 'o', 44,
		18, 't', 'w', 'o', '-', 't', 'h', 'r', 'e', 'e', 46,
		0, // End of blocks

		0, // End of blocks
	}

	benchmarkMapSchema := `{
		"name": "NestedMapRecord",
		"type": "record",
		"fields": [{
			"name": "aInt",
			"type": "int"
		}, {
			"name": "optInt",
			"type": ["null", "int"]
		}, {
			"name": "nodes",
			"type": {
				"type": "map",
				"values": {
					"name": "numberMap",
					"type":	"record",
					"fields": [{
						"name": "index",
						"type": "float"
					}, {
						"name": "numbers",
						"type": {
							"type": "map",
							"values": "int"
						}
					}]
				}
			}
		}]
	}`

	p, err := compiler.CompileSchemaBytes([]byte(benchmarkMapSchema), []byte(benchmarkMapSchema))
	require.Nil(b, err)

	engine := vm.Engine{
		Program: p,
	}

	for n := 0; n < b.N; n++ {
		var obj generated.NestedMapRecord
		err = engine.Run(bytes.NewBuffer(benchmarkMapInputData), &obj)
	}
	benchMapErr = err
}
