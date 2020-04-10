package vm

import (
	"fmt"
)

// Opcode defines the data type for an opcode
type Opcode byte

// Constant values for Opcode
const (
	OpError    Opcode = iota // Zero-value opcode is a bug in the program
	OpRet                    // returns from a subroutine
	OpHalt                   // Stops execution with an error (TODO implement error codes)
	OpSort                   // Instructs the VM to reorder the reader's fields when reading
	OpLoad                   // Loads a word from input into accumulator
	OpMov                    // Moves input data from the operand type tt to the current placeholder
	OpDiscard                // Discards as much data from input as required for type tt
	OpSkip                   // Skips reading the current field
	OpJmp                    // jumps to the relative position pp
	OpCase                   // jumps to the relative position pp if acc is equal to val
	OpSkipCase               // skips the current field and jumps to the relative position pp if acc is equal to val
	OpRecord                 // calls a subroutine for reading a record type
	OpBlock                  // Handles the loop for decoding blocks of data
	OpEndBlock               // Closes the innermost loop
)

func (op Opcode) String() string {
	switch op {
	case OpError:
		return "<error>"
	case OpRet:
		return "ret"
	case OpHalt:
		return "halt"
	case OpSort:
		return "sort"
	case OpLoad:
		return "load"
	case OpMov:
		return "mov"
	case OpDiscard:
		return "discard"
	case OpSkip:
		return "skip"
	case OpJmp:
		return "jmp"
	case OpCase:
		return "case"
	case OpSkipCase:
		return "skipCase"
	case OpRecord:
		return "record"
	case OpBlock:
		return "block"
	case OpEndBlock:
		return "blockEnd"
	default:
		return fmt.Sprintf("<invalid opCode %d>", op)
	}
}
