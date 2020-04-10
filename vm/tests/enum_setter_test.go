package tests

import (
	"bytes"
	"testing"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/generated"
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

	enumSchema = `{
		"name": "EnumTestRecord",
		"type": "record",
		"fields": [{
			"name": "aInt",
			"type": "int"
		}, {
			"name": "aEnum",
			"type": {
				"name": "numbers",
				"type":	"enum",
				"symbols": ["zero", "one", "two"]
			}
		}, {
			"name": "optEnum",
			"type": ["null", "numbers"]
		}]
	}`
)

func TestEnumSetter(t *testing.T) {
	p, err := compiler.CompileSchemaBytes([]byte(enumSchema), []byte(enumSchema))
	require.Nil(t, err)

	engine := vm.Engine{
		Program:     p,
		StackTraces: true,
	}

	var obj generated.SetterEnumRecord
	if err = engine.Run(bytes.NewBuffer(enumInputData), &obj); err != nil {
		t.Fatalf("Program failed: %v", err)
	}

	assert.Equal(t, int32(42), obj.AInt)
	assert.Equal(t, generated.EnumOne, obj.AEnum)
	require.NotNil(t, obj.OptEnum)
	assert.Equal(t, generated.EnumTwo, *obj.OptEnum)
}
