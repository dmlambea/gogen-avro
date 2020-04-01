package compiler

import (
	"container/list"

	"github.com/actgardner/gogen-avro/vm"
)

func (c *compiler) link(main *method) (insts []vm.Instruction) {
	methodList := list.New()

	main.offset = 0
	offset := main.Size()
	methodList.PushBack(main)

	// Compute method sizes
	for _, m := range c.methods {
		m.offset = offset
		offset += m.Size()
		methodList.PushBack(m)
	}

	// Compute method relative calls/jumps
	insts = make([]vm.Instruction, offset)
	for elem := methodList.Front(); elem != nil; elem = elem.Next() {
		m := elem.Value.(*method)
		pos := m.offset
		for i := range m.code {
			insts[pos] = m.code[i]
			// Block instructions' positions are already computed by the compiler, since their targets
			// belong to the same method they're getting compiled in.
			if insts[pos].IsJumpType() && !insts[pos].IsBlockType() {
				targetMethod := m.methodRefs[i]
				insts[pos].SetPos(targetMethod.offset - pos - 1)
			}
			pos++
		}
	}
	return
}
