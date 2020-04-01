package vm

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgram(t *testing.T) {
	instructionSet := []Instruction{
		Halt(),
		Load(),
		Skip(),
		EndBlock(),
		Ret(),

		// 2-byte instructions
		Mov(TypeInt),
		Jmp(5),
		Record(-4),
		Discard(TypeString),
		Block(6),

		// 3-byte instructions
		MovEq(1, TypeInt),
		JmpEq(1, 2),
		RecordEq(1, -4),
		DiscardBlock(-10),
		DiscardRecord(-10),
		DiscardEq(1, TypeString),
		BlockEq(1, 1),

		// 4-byte instructions
		DiscardEqBlock(1, -8),
		DiscardEqRecord(1, -8),

		// n-byte instructions
		Sort([]int{3, 2, 1, 0}),
	}

	p := NewProgram(instructionSet)
	var buf bytes.Buffer
	_, err := p.WriteTo(&buf)
	require.Nil(t, err)
	goldenEquals(t, "testProgram", buf.Bytes())

	p, err = NewProgramFromBytecode(buf.Bytes())
	require.Nil(t, err)
	assert.EqualValues(t, instructionSet, p.instructions)
}
