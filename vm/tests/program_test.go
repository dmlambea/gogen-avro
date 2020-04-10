package tests

import (
	"bytes"
	"testing"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgram(t *testing.T) {
	instructionSet := []vm.Instruction{
		vm.Ret(),
		vm.Halt(0),
		vm.Load(),
		vm.Skip(),
		vm.EndBlock(),

		// 2-byte instructions
		vm.Mov(vm.TypeAcc),
		vm.Jmp(5),
		vm.Record(-4),
		vm.Discard(vm.TypeString),
		vm.Block(6),

		// 3-byte instructions
		vm.Case(1, 2),
		vm.SkipCase(0, 3),
		vm.DiscardBlock(-10),
		vm.DiscardRecord(-10),

		// n-byte instructions
		vm.Sort([]int{3, 2, 1, 0}),
	}

	p := vm.NewProgram(instructionSet, nil)
	asm1 := p.String()
	var buf bytes.Buffer
	_, err := p.WriteTo(&buf)
	require.Nil(t, err)
	goldenEquals(t, "testProgram", buf.Bytes())

	p, err = vm.NewProgramFromBytecode(buf.Bytes())
	require.Nil(t, err)
	asm2 := p.String()
	assert.Equal(t, asm1, asm2)
}
