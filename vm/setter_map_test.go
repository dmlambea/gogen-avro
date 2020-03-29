package vm

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/actgardner/gogen-avro/vm/generated"
	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	complexMapInputData = []byte{
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

	complexMapReaderByteCode = []byte{
		byte(OpMov), byte(TypeInt),
		byte(OpMovOpt), 0, byte(TypeInt),
		byte(OpLoopStart), 3,
		byte(OpMov), byte(TypeString), // Outer Map key
		byte(OpCall), byte(2), // Call NestedMapRecord
		byte(OpLoopEnd),
		byte(OpHalt),
		byte(OpMov), byte(TypeFloat), // NestedMapRecord
		byte(OpLoopStart), 3,
		byte(OpMov), byte(TypeString), // Inner Map key
		byte(OpMov), byte(TypeInt), // Inner map values
		byte(OpLoopEnd),
		byte(OpRet),
	}

	simpleMapInputData = []byte{
		84,

		4, // Block length: 2 items

		8, 'k', 'e', 'y', '1',
		8, 'v', 'a', 'l', '1',

		8, 'k', 'e', 'y', '2',
		8, 'v', 'a', 'l', '2',

		0, // End of blocks
	}

	simpleMapReaderByteCode = []byte{
		byte(OpMov), byte(TypeInt),
		byte(OpLoopStart), 3,
		byte(OpMov), byte(TypeString), // Outer Map key
		byte(OpMov), byte(TypeString), // Outer Map value
		byte(OpLoopEnd),
		byte(OpHalt),
	}
)

func TestSimpleMapSetter(t *testing.T) {
	p, err := NewProgram(simpleMapReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SimpleMapTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(p, objSetter)
	err = engine.Run(bytes.NewBuffer(simpleMapInputData))
	if err != nil {
		t.Fatalf("Program failed:\n%s\n\nFailure: %v", p.String(), err)
	}

	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, int32(42), obj.AInt)

	require.Equal(t, 2, len(obj.Nodes))

	var value generated.Nested
	var ok bool

	value, ok = obj.Nodes["key1"]
	assert.True(t, ok)
	assert.Equal(t, "val1", value.Val)

	value, ok = obj.Nodes["key2"]
	assert.True(t, ok)
	assert.Equal(t, "val2", value.Val)
}

func TestMapSetter(t *testing.T) {
	p, err := NewProgram(complexMapReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterMapTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(p, objSetter)
	err = engine.Run(bytes.NewBuffer(complexMapInputData))
	fmt.Printf("Obj:\t%+v\n", obj)
	if err != nil {
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

var mainErr error

func BenchmarkMapSetter(b *testing.B) {
	var err error
	for n := 0; n < b.N; n++ {
		p, _ := NewProgram(complexMapReaderByteCode)

		var obj generated.SetterMapTestRecord
		objSetter, _ := setters.NewSetterFor(&obj)

		engine := NewEngine(p, objSetter)
		err = engine.Run(bytes.NewBuffer(complexMapInputData))
	}
	mainErr = err
}
