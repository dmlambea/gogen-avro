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
	enumInputData = []byte{
		84, // AInt
		2,  // AEnum "one"
		2,  // OptEnum pointer (valid)
		4,  // OptEnum "two"
	}

	enumReaderByteCode = []byte{
		byte(OpMov), byte(TypeInt),
		byte(OpMov), byte(TypeInt),
		byte(OpMovEq), 1, byte(TypeInt),
		byte(OpHalt),
	}
)

func TestEnumSetter(t *testing.T) {
	p, err := NewProgram(enumReaderByteCode)
	assert.Nil(t, err)

	var obj generated.SetterEnumRecord
	objSetter, err := setters.NewSetterFor(&obj)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(p, objSetter)
	err = engine.Run(bytes.NewBuffer(enumInputData))
	if err != nil {
		t.Fatalf("Program failed:\n%s\n\nFailure: %v", p.String(), err)
	}

	assert.Equal(t, int32(42), obj.AInt)
	assert.Equal(t, generated.EnumOne, obj.AEnum)
	require.NotNil(t, obj.OptEnum)
	assert.Equal(t, generated.EnumTwo, *obj.OptEnum)
	fmt.Printf("Result: %+v\n", obj)
}
