package vm

import (
	"fmt"
	"io"
	"strings"
)

// NewProgram creates a runnable program from the given instruction list
func NewProgram(instructions []Instruction, errorMessages []string) Program {
	return Program{instructions: instructions, errors: errorMessages}
}

// NewProgramFromBytecode compiles bytecode onto a runnable program
func NewProgramFromBytecode(byteCode []byte) (Program, error) {
	p := Program{}
	for i, pos := 0, 0; i < len(byteCode); {
		inst := decodeInstruction(byteCode[i:])
		if inst.op == OpError {
			return p, fmt.Errorf("bad instruction %#v at byteCode pos %d", inst, i)
		}
		p.instructions = append(p.instructions, inst)
		i += inst.Size()
		pos++
	}
	return p, nil
}

// Program holds the instruction and error codes for a given program
type Program struct {
	// The list of instructions that make up the deserializer program
	instructions []Instruction

	// A list of errors that can be triggered by halt(x), where x is the index in this array + 1
	errors []string
}

// WriteTo implements io.Writer for this Program
func (p Program) WriteTo(w io.Writer) (int64, error) {
	var sum int64 = 0
	for i := range p.instructions {
		n, err := p.instructions[i].WriteTo(w)
		sum += n
		if err != nil {
			return sum, err
		}
	}
	return sum, nil
}

func (p Program) String() string {
	var b strings.Builder
	for i, inst := range p.instructions {
		b.WriteString(fmt.Sprintf("%d:\t%s", i, inst.String()))
		if inst.IsJumpType() {
			b.WriteString(fmt.Sprintf(" [%d]", i+inst.pos+1))
		}
		b.WriteString("\n")
	}

	for i, err := range p.errors {
		b.WriteString(fmt.Sprintf("err %d:\t%s\n", i, err))
	}
	return b.String()
}
