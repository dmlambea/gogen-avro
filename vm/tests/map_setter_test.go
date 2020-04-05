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
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpMovEq), 0, byte(vm.TypeInt),
		byte(vm.OpBlock), 3,
		byte(vm.OpMov), byte(vm.TypeString), // Outer Map key
		byte(vm.OpRecord), byte(2), // Call NestedMapRecord
		byte(vm.OpEndBlock),
		byte(vm.OpRet),
		byte(vm.OpMov), byte(vm.TypeFloat), // NestedMapRecord
		byte(vm.OpBlock), 3,
		byte(vm.OpMov), byte(vm.TypeString), // Inner Map key
		byte(vm.OpMov), byte(vm.TypeInt), // Inner map values
		byte(vm.OpEndBlock),
		byte(vm.OpRet),
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
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpBlock), 3,
		byte(vm.OpMov), byte(vm.TypeString), // Outer Map key
		byte(vm.OpMov), byte(vm.TypeString), // Outer Map value
		byte(vm.OpEndBlock),
		byte(vm.OpHalt),
	}
)

func TestSimpleMapSetter(t *testing.T) {
	p, err := vm.NewProgramFromBytecode(simpleMapReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SimpleMapTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := vm.NewEngine(p, objSetter)
	if err = engine.Run(bytes.NewBuffer(simpleMapInputData)); err != nil {
		t.Fatalf("Program failed: %v", err)
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
	p, err := vm.NewProgramFromBytecode(complexMapReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterMapTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := vm.NewEngine(p, objSetter)
	if err = engine.Run(bytes.NewBuffer(complexMapInputData)); err != nil {
		t.Fatalf("Program failed: %v", err)
	}
	t.Logf("Result: %+v\n", obj)

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
		p, _ := vm.NewProgramFromBytecode(complexMapReaderByteCode)

		var obj generated.SetterMapTestRecord
		objSetter, _ := setters.NewSetterFor(&obj)

		engine := vm.NewEngine(p, objSetter)
		err = engine.Run(bytes.NewBuffer(complexMapInputData))
	}
	mainErr = err
}
