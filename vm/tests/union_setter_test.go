package tests

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/generated"
	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unionNullStringIntNodeFixture struct {
	input    []byte
	expected generated.UnionStringIntNode
}

var (
	unionNullStringIntNodeSchema = `["null", "string", "int", {
		"name": "node",
		"type": "record",
		"fields": [{
			"name": "name",
			"type": "string"
		}, {
			"name": "addr",
			"type": {
				"name": "address",
				"type": "record",
				"fields": [{
					"name": "id",
					"type": "int"
				}, {
					"name": "next",
					"type": ["null", "address"]
				}]
			}
		}]
	}]
	`

	unionNullStringIntNodeFixtures = []unionNullStringIntNodeFixture{
		{input: []byte{0}, expected: generated.UnionStringIntNode{}},
		{input: []byte{2, 8, 'T', 'e', 's', 't'}, expected: generated.UnionStringIntNode{setters.BaseUnion{Type: 1, Value: "Test"}}},
		{input: []byte{4, 84}, expected: generated.UnionStringIntNode{setters.BaseUnion{Type: 2, Value: int32(42)}}},
		{input: []byte{6, 12, 'N', 'o', 'd', 'e', '-', '1', 2, 0}, expected: generated.UnionStringIntNode{
			setters.BaseUnion{
				3,
				generated.Node{
					Name: "Node-1",
					Addr: &generated.Address{Id: 1},
				},
			}},
		},
	}
)

func TestUnion(t *testing.T) {
	p, err := compiler.CompileSchemaBytes([]byte(unionNullStringIntNodeSchema), []byte(unionNullStringIntNodeSchema))
	require.Nil(t, err)

	engine := vm.Engine{
		Program:     p,
		StackTraces: true,
	}

	for i, f := range unionNullStringIntNodeFixtures {
		var obj generated.UnionStringIntNode

		buf := bytes.NewBuffer(f.input)
		err = engine.Run(buf, &obj)
		require.Nil(t, err)

		assert.Equal(t, f.expected, obj, fmt.Sprintf("Union %d fails", i))
	}
}

func TestUnionError(t *testing.T) {
	p, err := compiler.CompileSchemaBytes([]byte(unionNullStringIntNodeSchema), []byte(unionNullStringIntNodeSchema))
	require.Nil(t, err)

	engine := vm.Engine{
		Program:     p,
		StackTraces: false,
	}

	var obj generated.UnionStringIntNode

	buf := bytes.NewBuffer([]byte{8}) // 8 is zigzag-encoded value for 4
	err = engine.Run(buf, &obj)
	require.NotNil(t, err)
	assert.Equal(t, "execution halted: invalid index for union", err.Error(), "bad error message")
}

type sortedUnion struct {
	setters.BaseUnion
}

func (u sortedUnion) UnionTypes() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*float64)(nil)),
		reflect.TypeOf((*bool)(nil)),
	}
}

func TestSortedUnion(t *testing.T) {
	writerSchemas := []string{
		`["boolean", "int"]`,
		`["boolean", "int"]`,
		`["boolean", "int"]`,
		`["boolean", "int"]`,
	}

	readerSchemas := []string{
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aInt", "type": "int"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aLong", "type": "long"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aFloat", "type": "float"}, { "name": "aBool", "type": "boolean"} ] }`,
		`{ "type": "record", "name": "TestRec", "fields": [ { "name": "aDouble", "type": "double"}, { "name": "aBool", "type": "boolean"} ] }`,
	}

	var obj sortedUnion

	for i := range writerSchemas {
		p, err := compiler.CompileSchemaBytes([]byte(writerSchemas[i]), []byte(readerSchemas[i]))
		require.Nil(t, err)

		engine := vm.Engine{
			Program:     p,
			StackTraces: true,
		}

		buf := bytes.NewBuffer([]byte{02, 84}) // Writer's second field of the union, value 42
		err = engine.Run(buf, &obj)
		require.Nil(t, err)

		assert.Equal(t, obj.Type, int64(1), fmt.Sprintf("Union %d: wrong union type %d", i, obj.Type))
		assert.Equal(t, obj.Value, int32(42), fmt.Sprintf("Union %d: wrong union value %v", i, obj.Value))
	}
}
