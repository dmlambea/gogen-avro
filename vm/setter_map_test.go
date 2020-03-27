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
	mapInputData = []byte{
		84,
		0, 42,

		4, // Block length: 2 items

		10, 'd', 'e', 'm', 'o', '1',

		00, 00, 12, 34,
		2, // Block length: 1 item
		12, 'q', 'u', 'i', 'n', 'c', 'e', 30,
		0, // End of blocks

		10, 'd', 'e', 'm', 'o', '2',

		00, 00, 23, 45,
		6, // Block length: 3 item
		10, 'c', 'i', 'n', 'c', 'o', 10,
		8, 's', 'e', 'i', 's', 12,
		10, 's', 'i', 'e', 't', 'e', 14,
		0, // End of blocks

		0, // End of blocks
	}

	mapReaderByteCode = []byte{
		byte(OpMov), byte(TypeInt),
		byte(OpMovOpt), 0, byte(TypeInt),
		byte(OpLoopStart), 7,
		byte(OpMov), byte(TypeString), // Outer Map key
		byte(OpMov), byte(TypeFloat), // Outer map values
		byte(OpLoopStart), 3,
		byte(OpMov), byte(TypeString), // Inner Map key
		byte(OpMov), byte(TypeInt), // Inner map values
		byte(OpLoopEnd),
		byte(OpLoopEnd),
		byte(OpHalt),
	}
)

func TestMapSetter(t *testing.T) {
	p, err := NewProgram(mapReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterMapTestRecord
	objSetter, err := setter.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(p, objSetter)
	err = engine.Run(bytes.NewBuffer(mapInputData))
	if err != nil && err != setter.ErrSetterEOF {
		t.Fatalf("Program failed:\n%s\n\nFailure: %v", p.String(), err)
	}

	assert.Equal(t, int32(42), obj.AInt)

	require.NotNil(t, obj.OptInt)
	assert.Equal(t, int32(21), *obj.OptInt)

	require.Equal(t, 2, len(obj.Nodes))

	var r generated.NestedMapRecord
	var ok bool
	r, ok = obj.Nodes["demo1"]
	assert.True(t, ok)
	assert.True(t, r.Index != 0)
	require.Equal(t, 1, len(r.Numbers))
	assert.Equal(t, int32(15), r.Numbers["quince"])

	r, ok = obj.Nodes["demo2"]
	assert.True(t, ok)
	assert.True(t, r.Index != 0)
	require.Equal(t, 3, len(r.Numbers))
	assert.Equal(t, int32(5), r.Numbers["cinco"])
	assert.Equal(t, int32(6), r.Numbers["seis"])
	assert.Equal(t, int32(7), r.Numbers["siete"])
}
