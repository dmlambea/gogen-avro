package compiler

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/stretchr/testify/require"
)

type complexTypeRoundtripFixture struct {
	index     int
	name      string
	wrtSchema []byte
	inputData []byte
	target    interface{}
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
	var obj1 complexTypeEnumIntRecString
	var obj2 complexTypeFloatRecsliceStringRecBytes
	targets := []interface{}{
		&obj1,
		&obj2,
	}

	f := complexTypeRoundtripFixture{name: "complex_struct"}
	f.wrtSchema = goldenLoad(f.name + "_writer.avsc")
	f.inputData = goldenLoad(f.name + "_data.bin")

	for i := range targets {
		f.index = i
		f.target = targets[i]
		roundtrip(t, f)
	}
}

func roundtrip(t *testing.T, f complexTypeRoundtripFixture) {
	readerSchema := goldenLoad(fmt.Sprintf("%s_reader_%d.avsc", f.name, f.index))

	p, err := CompileSchemaBytes(f.wrtSchema, readerSchema)
	require.Nil(t, err)
	goldenEquals(t, fmt.Sprintf("%s_%d.asm", f.name, f.index), []byte(p.String()))

	engine := vm.Engine{
		Program:     p,
		StackTraces: true,
	}
	err = engine.Run(bytes.NewBuffer(f.inputData), f.target)
	require.Nil(t, err)
	goldenEquals(t, fmt.Sprintf("%s_%d.out", f.name, f.index), []byte(fmt.Sprintf("%+v", f.target)))
}
