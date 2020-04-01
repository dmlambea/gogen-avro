package vm

import (
	"fmt"
)

// Opcode defines the data type for an opcode
type Opcode byte

// Constant values for Opcode
const (
	OpError     Opcode = iota // Zero-value opcode is a bug in the program
	OpHalt                    // Stops execution with an error (TODO implement error codes)
	OpSort                    // Instructs the VM to reorder the reader's fields when reading
	OpLoad                    // Loads a word from input into accumulator
	OpMov                     // Moves input data from the operand type tt to the current placeholder
	OpMovEq                   // Executes Load and then executes Mov if the acc is equal to val
	OpDiscard                 // Discards as much data from input as required for type tt
	OpDiscardEq               // Discards as much data from input as required for type tt if acc is equals to val
	OpSkip                    // Skips reading the current field
	OpJmp                     // jumps to the relative position pp
	OpJmpEq                   // jumps to the relative position pp if acc is equal to val
	OpRecord                  // calls a subroutine for reading a record type
	OpRecordEq                // calls a subroutine for reading a record type if acc is equals to val
	OpRet                     // returns from a subroutine
	OpBlock                   // Handles the loop for decoding blocks of data
	OpBlockEq                 // Handles the loop for decoding blocks of data if acc is equals to val
	OpEndBlock                // Closes the innermost loop
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
	case OpRecord:
		return "record"
	case OpRecordEq:
		return "recordEq"
	case OpRet:
		return "ret"
	case OpBlock:
		return "block"
	case OpEndBlock:
		return "blockEnd"
	default:
		return fmt.Sprintf("<invalid opCode %d>", op)
	}
}
