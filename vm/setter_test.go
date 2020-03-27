package vm

import (
	"bytes"
	"testing"

	"github.com/actgardner/gogen-avro/vm/generated"
	"github.com/actgardner/gogen-avro/vm/setter"
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

		// Node :: Address (valid)
		2,
		2, // Id

		// Next :: Address (valid)
		2,
		4, // Id

		// Next :: Address (valid)
		2,
		6, // Id

		// Next :: Address (nil)
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
		byte(OpCall), 6,
		byte(OpHalt),
		byte(OpMov), byte(TypeString), // 4: Node reader
		byte(OpLoad),
		byte(OpJmpEq), 0, 1,
		byte(OpCall), 1,
		byte(OpRet),
		byte(OpMov), byte(TypeInt), // Address reader
		byte(OpLoad),
		byte(OpJmpEq), 0, 1,
		byte(OpJmp), -4 & 0xff,
		byte(OpRet),
	}
)

func TestSetter(t *testing.T) {
	p, err := NewProgram(readerByteCode)
	assert.Nil(t, err)

	var obj generated.SetterTestRecord
	objSetter, err := setter.NewSetterFor(&obj)
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

	require.NotNil(t, obj.Node.Addr.Next.Next.Next)
	assert.Equal(t, int32(4), obj.Node.Addr.Next.Next.Next.Id)

	assert.Nil(t, obj.Node.Addr.Next.Next.Next.Next)

	require.NotNil(t, obj.OptAddr)
	assert.Equal(t, int32(5), obj.OptAddr.Id)
	assert.Nil(t, obj.OptAddr.Next)
}
