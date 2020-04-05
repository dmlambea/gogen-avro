package tests

import (
	"bytes"
	"testing"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/generated"
	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnionString(t *testing.T) {
	var obj generated.UnionStringIntNode

	objSetter, err := setters.NewSetterFor(&obj)
	require.Nil(t, err)
	require.NotNil(t, objSetter)

	p := vm.NewProgram([]vm.Instruction{
		vm.Mov(vm.TypeLong),
		vm.Mov(vm.TypeString),
		vm.Ret(),
	})

	var buf bytes.Buffer
	buf.WriteByte(2) // Union type 1
	buf.WriteByte(6) // String len
	buf.WriteString("Hi!")

	engine := vm.NewEngine(p, objSetter)
	err = engine.Run(&buf)
	require.Nil(t, err)
	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, generated.UnionStringIntNodeTypeString, obj.Type)
	assert.Equal(t, "Hi!", obj.Value)
}

func TestUnionInt(t *testing.T) {
	var obj generated.UnionStringIntNode

	objSetter, err := setters.NewSetterFor(&obj)
	require.Nil(t, err)
	require.NotNil(t, objSetter)

	p := vm.NewProgram([]vm.Instruction{
		vm.Mov(vm.TypeLong),
		vm.Mov(vm.TypeInt),
		vm.Ret(),
	})

	var buf bytes.Buffer
	buf.WriteByte(4)  // Union type 2
	buf.WriteByte(84) // Int 42

	engine := vm.NewEngine(p, objSetter)
	err = engine.Run(&buf)
	require.Nil(t, err)
	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, generated.UnionStringIntNodeTypeInt, obj.Type)
	assert.Equal(t, int32(42), obj.Value)
}

func TestUnionNode(t *testing.T) {
	var obj generated.UnionStringIntNode

	objSetter, err := setters.NewSetterFor(&obj)
	require.Nil(t, err)
	require.NotNil(t, objSetter)

	p := vm.NewProgram([]vm.Instruction{
		vm.Mov(vm.TypeLong),
		vm.Record(1),
		vm.Ret(),
		vm.Mov(vm.TypeString),
		vm.RecordEq(1, 1),
		vm.Ret(),
		vm.Mov(vm.TypeInt),
		vm.RecordEq(1, -2),
		vm.Ret(),
	})

	var buf bytes.Buffer
	buf.WriteByte(6) // Union type 3
	buf.WriteByte(6) // String length
	buf.WriteString("Hi!")
	buf.WriteByte(2)  // Opt address follows
	buf.WriteByte(42) // Address ID 21
	buf.WriteByte(0)  // Opt Next

	engine := vm.NewEngine(p, objSetter)
	err = engine.Run(&buf)
	require.Nil(t, err)
	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, generated.UnionStringIntNodeTypeNode, obj.Type)
	assert.Equal(t, "Hi!", obj.Value.(generated.Node).Name)
	require.NotNil(t, obj.Value.(generated.Node).Addr)
	assert.Equal(t, int32(21), obj.Value.(generated.Node).Addr.Id)
	assert.Nil(t, obj.Value.(generated.Node).Addr.Next)
}
