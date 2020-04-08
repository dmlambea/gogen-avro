package compiler

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/require"
)

type complexTypeObjectSetterPair struct {
	object interface{}
	setter setters.Setter
}

type complexTypeRoundtripFixture struct {
	index     int
	name      string
	wrtSchema []byte
	inputData []byte
	target    complexTypeObjectSetterPair
}

type complexTypeEnumIntRecString struct {
	Enum int32
	AInt int32
	ARec struct {
		InnerInt int32
	}
	Text string
}

type complexTypeFloatRecsliceStringRecBytes struct {
	AFloat    float32
	ARecArray []struct {
		InnerInt int32
	}
	AString string
	ARec    struct {
		InnerInt int32
	}
	ABytes []byte
}

func TestCompileComplex(t *testing.T) {
	setterOrFail := func(s setters.Setter, err error) setters.Setter {
		require.Nil(t, err)
		return s
	}

	var obj1 complexTypeEnumIntRecString
	var obj2 complexTypeFloatRecsliceStringRecBytes
	pairs := []complexTypeObjectSetterPair{
		{object: &obj1, setter: setterOrFail(setters.NewSetterFor(&obj1))},
		{object: &obj2, setter: setterOrFail(setters.NewSetterFor(&obj2))},
	}

	f := complexTypeRoundtripFixture{name: "complex_struct"}
	f.wrtSchema = goldenLoad(f.name + "_writer.avsc")
	f.inputData = goldenLoad(f.name + "_data.bin")

	for i := range pairs {
		f.index = i
		f.target = pairs[i]
		roundtrip(t, f)
	}
}

func roundtrip(t *testing.T, f complexTypeRoundtripFixture) {
	readerSchema := goldenLoad(fmt.Sprintf("%s_reader_%d.avsc", f.name, f.index))

	prog, err := CompileSchemaBytes(f.wrtSchema, readerSchema)
	require.Nil(t, err)
	goldenEquals(t, fmt.Sprintf("%s_%d.asm", f.name, f.index), []byte(prog.String()))

	engine := vm.NewEngine(prog, f.target.setter)
	err = engine.Run(bytes.NewBuffer(f.inputData))
	require.Nil(t, err)
	goldenEquals(t, fmt.Sprintf("%s_%d.out", f.name, f.index), []byte(fmt.Sprintf("%+v", f.target.object)))
}
