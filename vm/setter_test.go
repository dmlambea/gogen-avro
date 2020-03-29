package vm

import (
	"bytes"
	"testing"

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
		byte(OpMov), byte(TypeInt),
		byte(OpMovOpt), 0, byte(TypeInt),
		byte(OpMovOpt), 0, byte(TypeInt),
		byte(OpCall), 4,
		byte(OpLoad),
		byte(OpJmpEq), 0, 1,
		byte(OpCall), 8,
		byte(OpHalt),
		// Node reader
		byte(OpMov), byte(TypeString),
		byte(OpLoad),
		byte(OpJmpEq), 1, 2,
		byte(OpSkip),
		byte(OpRet),
		byte(OpCall), 1,
		byte(OpRet),
		// Address reader
		byte(OpMov), byte(TypeInt),
		byte(OpLoad),
		byte(OpJmpEq), 0, 2,
		byte(OpCall), -4 & 0xff,
		byte(OpRet),
		byte(OpSkip),
		byte(OpRet),
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
		byte(OpSort), 2, 3, 0,
		byte(OpCall), 5,
		byte(OpMov), byte(TypeInt),
		byte(OpSkip),
		byte(OpSkip),
		byte(OpSkip),
		byte(OpHalt),
		// Node reader
		byte(OpMov), byte(TypeString),
		byte(OpLoad),
		byte(OpJmpEq), 1, 2,
		byte(OpSkip),
		byte(OpRet),
		byte(OpCall), 1,
		byte(OpRet),
		// Address reader
		byte(OpMov), byte(TypeInt),
		byte(OpLoad),
		byte(OpJmpEq), 0, 2,
		byte(OpCall), -4 & 0xff,
		byte(OpRet),
		byte(OpSkip),
		byte(OpRet),
	}
)

func TestSetter(t *testing.T) {
	p, err := NewProgram(readerByteCode)
	assert.Nil(t, err)

	var obj generated.SetterTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(p, objSetter)
	err = engine.Run(bytes.NewBuffer(inputData))
	if err != nil {
		t.Fatalf("Program failed:\n%s\n\nFailure: %v", p.String(), err)
	}

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
	p, err := NewProgram(reorderedReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(p, objSetter)
	err = engine.Run(bytes.NewBuffer(reorderedInputData))
	if err != nil {
		t.Fatalf("Program failed:\n%s\n\nFailure: %v", p.String(), err)
	}

	assert.Equal(t, int32(42), obj.AInt)
	assert.Nil(t, obj.OptInt)
	assert.Nil(t, obj.NilInt)
	assert.Equal(t, "First node", obj.Node.Name)
	require.NotNil(t, obj.Node.Addr)
	assert.Equal(t, int32(1), obj.Node.Addr.Id)
	assert.Nil(t, obj.Node.Addr.Next)
	assert.Nil(t, obj.OptAddr)
}
