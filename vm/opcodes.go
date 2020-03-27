package vm

import (
	"fmt"
)

// Opcode defines the data type for an opcode
type Opcode byte

// Constant values for Opcode
const (
	OpError     Opcode = iota // Zero-value opcode is a bug in the program
	OpHalt                    // Stops execution (TODO with exit code xx)
	OpLoad                    // Loads a word from input into accumulator
	OpMov                     // Moves input data from the operand type tt to the current placeholder
	OpMovOpt                  // Executes Load and then executes Mov if the acc is equal to val
	OpSkip                    // Skips input data from the operand type tt
	OpJmp                     // jumps to the relative position pp
	OpJmpEq                   // jumps to the relative position pp if acc is equal to val
	OpCall                    // calls a subroutine
	OpRet                     // returns from a subroutine
	OpLoopStart               // Handles the loop for decoding of blocks
	OpLoopEnd                 // Closes the innermost loop
)

func (op Opcode) String() string {
	switch op {
	case OpError:
		return "<error>"
	case OpHalt:
		return "halt"
	case OpLoad:
		return "load"
	case OpMov:
		return "mov"
	case OpMovOpt:
		return "movOpt"
	case OpSkip:
		return "skip"
	case OpJmp:
		return "jmp"
	case OpJmpEq:
		return "jmpEq"
	case OpCall:
		return "call"
	case OpRet:
		return "ret"
	case OpLoopStart:
		return "loopStart"
	case OpLoopEnd:
		return "loopEnd"
	default:
		return fmt.Sprintf("<invalid opCode %d>", op)
	}
}
