package vm

import (
	"errors"
	"fmt"
)

// Instruction defines a VM executable instruction
type Instruction struct {
	op  Opcode
	tp  Type
	val interface{}
	pos int
}

var (
	errEOF = errors.New("EOF")
)

// Size returns the serialized size of this instruction.
func (i Instruction) Size() int {
	switch i.op {
	case OpError, OpHalt, OpLoad, OpSkip, OpRet, OpLoopEnd:
		return 1
	case OpMov, OpDiscard, OpJmp, OpCall, OpLoopStart:
		return 2
	case OpMovOpt, OpJmpEq:
		return 3
	case OpSort:
		data := i.val.([]int)
		return 2 + len(data)
	default:
		panic(fmt.Sprintf("invalid instruction opCode %x", i.op))
	}
}

// decodeInstruction returns the instruction at position 0 of input. No error checking is performed to
// ensure the input is large enough for holding the entire instruction.
func decodeInstruction(input []byte) (inst Instruction) {
	switch Opcode(input[0]) {
	case OpHalt:
		inst = Halt()
	case OpSort:
		l := int(input[1])
		data := make([]int, l)
		for i := 0; i < l; i++ {
			data[i] = int(input[2+i])
		}
		inst = Sort(data)
	case OpLoad:
		inst = Load()
	case OpMov:
		inst = Mov(Type(input[1]))
	case OpMovOpt:
		inst = MovOpt(int(input[1]), Type(input[2]))
	case OpDiscard:
		inst = Discard(Type(input[1]))
	case OpSkip:
		inst = Skip()
	case OpJmp:
		inst = Jmp(input[1])
	case OpJmpEq:
		inst = JmpEq(int(input[1]), input[2])
	case OpCall:
		inst = Call(input[1])
	case OpRet:
		inst = Ret()
	case OpLoopStart:
		inst = LoopStart(input[1])
	case OpLoopEnd:
		inst = LoopEnd()
	default:
		inst = Instruction{op: OpError}
	}
	return
}

func Halt() Instruction {
	return Instruction{op: OpHalt}
}

func Sort(data []int) Instruction {
	return Instruction{op: OpSort, val: data}
}

func Load() Instruction {
	return Instruction{op: OpLoad}
}

func Mov(t Type) Instruction {
	return Instruction{op: OpMov, tp: t}
}

func MovOpt(val int, t Type) Instruction {
	return Instruction{op: OpMovOpt, tp: t, val: val}
}

func Discard(t Type) Instruction {
	return Instruction{op: OpDiscard, tp: t}
}

func Skip() Instruction {
	return Instruction{op: OpSkip}
}

func Jmp(relByte byte) Instruction {
	return Instruction{op: OpJmp, pos: relByteToInt(relByte)}
}

func JmpEq(val int, relByte byte) Instruction {
	return Instruction{op: OpJmpEq, pos: relByteToInt(relByte), val: val}
}

func Call(relByte byte) Instruction {
	return Instruction{op: OpCall, pos: relByteToInt(relByte)}
}

func Ret() Instruction {
	return Instruction{op: OpRet}
}

func LoopStart(relByte byte) Instruction {
	return Instruction{op: OpLoopStart, pos: relByteToInt(relByte)}
}

func LoopEnd() Instruction {
	return Instruction{op: OpLoopEnd}
}

func relByteToInt(b byte) int {
	if b < 128 {
		return int(b)
	}
	return -(int(^b) + 1)
}

func (i Instruction) String() string {
	switch i.op {
	case OpError, OpHalt, OpLoad, OpSkip, OpRet, OpLoopEnd:
		return i.op.String()
	case OpMov:
		return fmt.Sprintf("%s %s", i.op, i.tp)
	case OpMovOpt:
		return fmt.Sprintf("%s %s %d", i.op, i.tp, i.val)
	case OpJmp, OpCall, OpLoopStart:
		return fmt.Sprintf("%s -> %d", i.op, i.pos)
	case OpJmpEq:
		return fmt.Sprintf("%s %d -> %d", i.op, i.val, i.pos)
	case OpSort:
		return fmt.Sprintf("%s %v", i.op, i.val)
	default:
		return fmt.Sprintf("<invalid opCode %d>", i.op)
	}
}
