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

var (
	inputData = []byte{
		84,    // AInt
		0, 42, // OptInt (valid)
		2, // NilInt (nil)

		// hidden is omitted

		// Node :: Name
		20, 70, 105, 114, 115, 116, 32, 110, 111, 100, 101,

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

	readerByteCode = []byte{
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpMovEq), 0, byte(vm.TypeInt),
		byte(vm.OpMovEq), 0, byte(vm.TypeInt),
		byte(vm.OpRecord), 2,
		byte(vm.OpRecordEq), 1, 4,
		byte(vm.OpRet),
		// Node reader
		byte(vm.OpMov), byte(vm.TypeString),
		byte(vm.OpRecordEq), 1, 1,
		byte(vm.OpRet),
		// Address reader
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpRecordEq), 1, -2 & 0xff,
		byte(vm.OpRet),
	}

	reorderedInputData = []byte{
		// Node :: Name
		20, 70, 105, 114, 115, 116, 32, 110, 111, 100, 101,

		// Node :: Address (valid)
		2,
		2, // Id

		// Next :: Address (nil)
		0,

		84, // AInt
	}

	reorderedReaderByteCode = []byte{
		byte(vm.OpSort), 2, 3, 0,
		byte(vm.OpRecord), 5,
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpSkip),
		byte(vm.OpSkip),
		byte(vm.OpSkip),
		byte(vm.OpRet),
		// Node reader
		byte(vm.OpMov), byte(vm.TypeString),
		byte(vm.OpRecordEq), 1, 1,
		byte(vm.OpRet),
		// Address reader
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpRecordEq), 1, -2 & 0xff,
		byte(vm.OpRet),
	}
)

func TestSetter(t *testing.T) {
	p, err := vm.NewProgramFromBytecode(readerByteCode)
	assert.Nil(t, err)

	var obj generated.SetterTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := vm.NewEngine(p, objSetter)
	if err = engine.Run(bytes.NewBuffer(inputData)); err != nil {
		t.Fatalf("Program failed: %v", err)
	}
	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, int32(42), obj.AInt)
	require.NotNil(t, obj.OptInt)
	assert.Equal(t, int32(21), *obj.OptInt)

	assert.Nil(t, obj.NilInt)

	assert.Equal(t, "First node", obj.Node.Name)

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
	p, err := vm.NewProgramFromBytecode(reorderedReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := vm.NewEngine(p, objSetter)
	if err = engine.Run(bytes.NewBuffer(reorderedInputData)); err != nil {
		t.Fatalf("Program failed: %v", err)
	}
	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, int32(42), obj.AInt)
	assert.Nil(t, obj.OptInt)
	assert.Nil(t, obj.NilInt)
	assert.Equal(t, "First node", obj.Node.Name)
	require.NotNil(t, obj.Node.Addr)
	assert.Equal(t, int32(1), obj.Node.Addr.Id)
	assert.Nil(t, obj.Node.Addr.Next)
	assert.Nil(t, obj.OptAddr)
}
