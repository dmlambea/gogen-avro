package vm

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const namePrefix = "array-"

type simpleArrayTestNode struct {
	Index int32
	Name  string
}

type simpleArrayTestRecord struct {
	Nodes []simpleArrayTestNode
}

func (c simpleArrayTestNode) validate() error {
	if !strings.HasPrefix(c.Name, namePrefix) {
		return fmt.Errorf("name does not start with '%s': %+v", namePrefix, c)
	}
	expected := fmt.Sprintf("%s%03d", namePrefix, c.Index)
	if c.Name != expected {
		return fmt.Errorf("expected name '%s', got '%s': %+v", expected, c.Name, c)
	}
	return nil
}

func getProgram() Program {
	return NewProgram([]Instruction{
		Block(2),
		Record(2),
		EndBlock(),
		Ret(),
		Mov(TypeInt),
		Mov(TypeString),
		Ret(),
	})
}

func getInputDataFor(amount, blocks int) *bytes.Buffer {
	current := 1
	var buf bytes.Buffer

	genBlock := func(items int) {
		buf.WriteByte(byte(items * 2)) // Binary-encoded integer 10
		for i := 0; i < items; i++ {
			buf.WriteByte(byte(current * 2)) // Binary-encoded int
			name := fmt.Sprintf("%s%03d", namePrefix, current)
			current++
			buf.WriteByte(byte(len(name) * 2))
			buf.WriteString(name)
		}
	}

	perBlock := amount / blocks
	remaining := amount % blocks

	for i := 0; i < blocks; i++ {
		genBlock(perBlock)
	}
	if remaining > 0 {
		genBlock(remaining)
	}
	genBlock(0)

	return &buf
}

func TestSimpleArrayRoundtrip(t *testing.T) {
	p := getProgram()

	var obj simpleArrayTestRecord
	objSetter, err := setters.NewSetterFor(&obj)
	require.Nil(t, err)

	engine := NewEngine(p, objSetter)
	err = engine.Run(getInputDataFor(10, 3))
	require.Nil(t, err)

	t.Logf("Result: %+v\n", obj)

	require.NotNil(t, obj.Nodes)
	assert.Equal(t, 10, len(obj.Nodes))

	for _, item := range obj.Nodes {
		assert.Nil(t, item.validate())
	}
}

var benchArrayErr error

func BenchmarkArraySetter(b *testing.B) {
	var err error
	for n := 0; n < b.N; n++ {
		p := getProgram()

		var obj simpleArrayTestRecord
		objSetter, _ := setters.NewSetterFor(&obj)

		engine := NewEngine(p, objSetter)
		err = engine.Run(getInputDataFor(10, 1))
	}
	mainErr = err
}
