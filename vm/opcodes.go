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
	OpSort                    // Instructs the VM to reorder the reader's fields when reading
	OpLoad                    // Loads a word from input into accumulator
	OpMov                     // Moves input data from the operand type tt to the current placeholder
	OpMovEq                   // Executes Load and then executes Mov if the acc is equal to val
	OpDiscard                 // Discards as much data from input as required for type tt
	OpDiscardEq               // Discards as much data from input as required for type tt
	OpSkip                    // Skips reading the current field
	OpJmp                     // jumps to the relative position pp
	OpJmpEq                   // jumps to the relative position pp if acc is equal to val
	OpCall                    // calls a subroutine
	OpCallEq                  // calls a subroutine
	OpRet                     // returns from a subroutine
	OpLoop                    // Handles the loop for decoding of blocks
	OpLoopEq                  // Handles the loop for decoding of blocks
	OpEndLoop                 // Closes the innermost loop
)

func (op Opcode) String() string {
	switch op {
	case OpError:
		return "<error>"
	case OpHalt:
		return "halt"
	case OpSort:
		return "sort"
	case OpLoad:
		return "load"
	case OpMov:
		return "mov"
	case OpMovEq:
		return "movEq"
	case OpDiscard:
		return "discard"
	case OpDiscardEq:
		return "discardEq"
	case OpSkip:
		return "skip"
	case OpJmp:
		return "jmp"
	case OpJmpEq:
		return "jmpEq"
	case OpCall:
		return "call"
	case OpCallEq:
		return "callEq"
	case OpRet:
		return "ret"
	case OpLoop:
		return "loop"
	case OpEndLoop:
		return "loopEnd"
	default:
		return fmt.Sprintf("<invalid opCode %d>", op)
	}
}
