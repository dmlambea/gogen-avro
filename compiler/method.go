package compiler

import (
	"fmt"
	"strings"

	"github.com/actgardner/gogen-avro/vm"
)

func newMethod(name string) *method {
	return &method{
		name:       name,
		methodRefs: make(map[int]*method),
	}
}

type method struct {
	name       string
	offset     int
	code       []vm.Instruction
	methodRefs map[int]*method
}

// Size returns the number of instructions of this method
func (m method) Size() int {
	return len(m.code)
}

func (m method) IsAnon() bool {
	return m.name == ""
}

func (m method) String() string {
	var buf strings.Builder
	if m.IsAnon() {
		buf.WriteString("<anon>")
	} else {
		buf.WriteString(m.name)
	}
	buf.WriteString(":\n")
	for idx, inst := range m.code {
		buf.WriteString(fmt.Sprintf("  %03d: %s", idx, inst.String()))
		if sub, ok := m.methodRefs[idx]; ok {
			buf.WriteString("\t --> ")
			buf.WriteString(sub.name)
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// record adds a record instruction to a method, and registers the calling instruction's
// relative position, so that the linker can efficiently locate it.
func (m *method) record(anotherMethod *method) int {
	return m.addJumpInstruction(vm.Record(0), anotherMethod)
}

// discardRecord adds a discard instruction for a method, and registers the calling instruction's
// relative position, so that the linker can efficiently locate it.
func (m *method) discardRecord(anotherMethod *method) int {
	return m.addJumpInstruction(vm.DiscardRecord(0), anotherMethod)
}

// block adds a block instruction and returns its position within the method.
// The linker would me much more inefficient matching block start/ends, so the fastest way is
// to compute the relative jump at compile time.
func (m *method) block() int {
	return m.addJumpInstruction(vm.Block(0), nil)
}

func (m *method) addJumpInstruction(inst vm.Instruction, anotherMethod *method) int {
	instPos := len(m.code)
	m.code = append(m.code, inst)
	if anotherMethod != nil {
		m.methodRefs[instPos] = anotherMethod
	}
	return instPos
}

func (m *method) append(instructions ...vm.Instruction) *method {
	m.code = append(m.code, instructions...)
	return m
}
