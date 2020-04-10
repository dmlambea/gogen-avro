package compiler

import (
	"github.com/actgardner/gogen-avro/vm"
)

const (
	switchJumpTypeUnresolved = -2
	switchJumpTypeEnd        = -1
)

func newSwitch() *switchBlock {
	return &switchBlock{}
}

type switchBlock struct {
	cases    []int64
	blocks   []*method
	codeSize int
}

func (sw *switchBlock) addCase(index int64, m *method) {
	sw.cases = append(sw.cases, index)
	sw.blocks = append(sw.blocks, m)
	sw.codeSize += m.Size() + 2 // the case-jump and the jump at the end of the code block
}

func (sw switchBlock) compileTo(m *method, errCode int) {
	// rel pos start counting from the next instruction, so this number goes past the halt
	totalCases := len(sw.cases)

	// Codesize accounts for each case's sel-jmp and codesize+jmp
	// The final block won't have any jmp, but the sel-block ends with a halt,
	// so the math is correct.
	insts := make([]vm.Instruction, totalCases+1)

	// Create the case-selection block
	accumBlockLen := 0
	for i := range sw.cases {
		insts[i] = vm.Case(sw.cases[i], totalCases-i+accumBlockLen)
		accumBlockLen += sw.blocks[i].Size() + 1 // The final jmp
	}
	insts[totalCases] = vm.Halt(int64(errCode))
	m.append(insts...)

	// Fill in the code blocks with a termination jmp to the end of the switch-case
	pos := 0
	absPos := len(m.code)
	endPos := sw.codeSize - totalCases - 2
	for i := range sw.blocks {
		// Copy code and method refs
		sw.blocks[i].appendTo(m)
		for blockPos, blockMethod := range sw.blocks[i].methodRefs {
			// Fix absoute positions of the method calls within this method
			m.methodRefs[blockPos+absPos+pos] = blockMethod
		}

		pos += sw.blocks[i].Size()

		if i < totalCases-1 {
			m.append(vm.Jmp(endPos - pos))
			pos++
		}
	}
}
