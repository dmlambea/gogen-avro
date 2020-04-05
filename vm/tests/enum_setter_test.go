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
	enumInputData = []byte{
		84, // AInt
		2,  // AEnum "one"
		2,  // OptEnum pointer (valid)
		4,  // OptEnum "two"
	}

	enumReaderByteCode = []byte{
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpMov), byte(vm.TypeInt),
		byte(vm.OpMovEq), 1, byte(vm.TypeInt),
		byte(vm.OpRet),
	}
)

func TestEnumSetter(t *testing.T) {
	p, err := vm.NewProgramFromBytecode(enumReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterEnumRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := vm.NewEngine(p, objSetter)
	if err = engine.Run(bytes.NewBuffer(enumInputData)); err != nil {
		t.Fatalf("Program failed: %v", err)
	}
	t.Logf("Result: %+v\n", obj)

	assert.Equal(t, int32(42), obj.AInt)
	assert.Equal(t, generated.EnumOne, obj.AEnum)
	require.NotNil(t, obj.OptEnum)
	assert.Equal(t, generated.EnumTwo, *obj.OptEnum)
}
