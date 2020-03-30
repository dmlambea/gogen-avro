package vm

import (
	"fmt"
	"io"
	"strings"
)

// NewProgram compiles bytecode onto a runnable program
func NewProgram(byteCode []byte) (Program, error) {
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

type Program struct {
	// The list of instructions that make up the deserializer program
	instructions []Instruction

	// A list of errors that can be triggered by halt(x), where x is the index in this array + 1
	errs []string
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
		b.WriteString(fmt.Sprintf("%d:\t%s\n", i, inst.String()))
	}

	for i, err := range p.errs {
		b.WriteString(fmt.Sprintf("Error %v:\t%v\n", i+1, err))
	}
	return b.String()
}
