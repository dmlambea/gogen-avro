package vm

import (
	"fmt"
	"strings"
)

// Compile bytecode onto a runnable program
func NewProgram(byteCode []byte) (Program, error) {
	p := Program{}
	for i, pos := 0, 0; i < len(byteCode); {
		inst := decodeInstruction(byteCode[i:])
		switch inst.op {
		case OpError:
			return p, fmt.Errorf("bad instruction %#v at byteCode pos %d", inst, i)
		case OpJmp, OpJmpEq, OpCall, OpLoopStart:
			// Jumps are always relative to the instruction following the current,
			// so they have to have added the current abs pos, plus one
			inst.pos = inst.pos + pos + 1
			if inst.pos < 0 {
				return p, fmt.Errorf("bad jmp %#v at byteCode pos %d", inst, i)
			}
		}
		p.instructions = append(p.instructions, inst)
		i += inst.Size()
		pos++
	}
	return p, nil
}

type Program struct {
	// The list of instructions that make up the deserializer program
	instructions []Instruction

	// A list of errors that can be triggered by halt(x), where x is the index in this array + 1
	errs []string
}

func (p Program) String() string {
	var b strings.Builder
	for i, inst := range p.instructions {
		b.WriteString(fmt.Sprintf("%d:\t%s\n", i, inst.String()))
	}

	for i, err := range p.errs {
		b.WriteString(fmt.Sprintf("Error %v:\t%v\n", i+1, err))
	}
	return b.String()
}
