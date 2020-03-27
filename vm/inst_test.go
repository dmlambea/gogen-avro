package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fixture struct {
	input  []byte
	expect Instruction
}

var (
	negativeJmp int = -10

	decodeFixtures = []fixture{
		fixture{
			input:  []byte{byte(OpError)},
			expect: Instruction{op: OpError},
		},

		fixture{
			input:  []byte{byte(OpHalt)},
			expect: Instruction{op: OpHalt},
		},

		fixture{
			input:  []byte{byte(OpLoad)},
			expect: Instruction{op: OpLoad},
		},

		fixture{
			input:  []byte{byte(OpMov), byte(TypeBool)},
			expect: Instruction{op: OpMov, tp: TypeBool},
		},
		fixture{
			input:  []byte{byte(OpMov), byte(TypeInt)},
			expect: Instruction{op: OpMov, tp: TypeInt},
		},

		fixture{
			input:  []byte{byte(OpMovOpt), 123, byte(TypeBool)},
			expect: Instruction{op: OpMovOpt, tp: TypeBool, val: 123},
		},
		fixture{
			input:  []byte{byte(OpMovOpt), 123, byte(TypeInt)},
			expect: Instruction{op: OpMovOpt, tp: TypeInt, val: 123},
		},

		fixture{
			input:  []byte{byte(OpJmp), 123},
			expect: Instruction{op: OpJmp, pos: 123},
		},
		fixture{
			input:  []byte{byte(OpJmp), byte(negativeJmp)},
			expect: Instruction{op: OpJmp, pos: negativeJmp},
		},

		fixture{
			input:  []byte{byte(OpJmpEq), 1, 123},
			expect: Instruction{op: OpJmpEq, pos: 123, val: 1},
		},
		fixture{
			input:  []byte{byte(OpJmpEq), 1, byte(negativeJmp)},
			expect: Instruction{op: OpJmpEq, pos: negativeJmp, val: 1},
		},

		fixture{
			input:  []byte{byte(OpSkip)},
			expect: Instruction{op: OpSkip},
		},
	}
)

func TestInstructionDecoder(t *testing.T) {
	for _, f := range decodeFixtures {
		inst := decodeInstruction(f.input)
		assert.Exactly(t, inst, f.expect)
	}
}
