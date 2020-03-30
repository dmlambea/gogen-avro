package vm

import (
	"fmt"
	"io"
)

// Instruction defines a VM executable instruction.
type Instruction struct {
	op  Opcode
	tp  Type
	val interface{}
	pos int
}

// Halt stops the VM.
// TODO: add exit code for erroring.
func Halt() Instruction {
	return Instruction{op: OpHalt}
}

// Sort tells the VM to rearrange the target struct's fields when reading data.
func Sort(data []int) Instruction {
	return Instruction{op: OpSort, val: data}
}

// Load reads an int64 from input into the accumulator.
func Load() Instruction {
	return Instruction{op: OpLoad}
}

// Mov tt loads data of type tt into current field.
func Mov(t Type) Instruction {
	return Instruction{op: OpMov, tp: t}
}

// MovEq vv tt loads data of type tt if the first int64 value of input matches vv.
// The int64 value is consumed this way and therefore discarded.
func MovEq(val int64, t Type) Instruction {
	return Instruction{op: OpMovEq, tp: t, val: val}
}

// Discard tt discards input data of type tt.
func Discard(t Type) Instruction {
	return Instruction{op: OpDiscard, tp: t}
}

// DiscardBlock xx is a special type of discard so that a full record of data can be
// discarded. Since the size of the data to be discarded cannot be known in advance,
// the VM needs to consume it completely. The argument xx is the relative address of
// the code that is used to consume the input.
func DiscardBlock(relPos int) Instruction {
	return Instruction{op: OpDiscard, tp: TypeBlock, pos: relPos}
}

// DiscardEq vv tt discards data of type tt, if the first int64 value of input
// matches vv. The int64 value is consumed this way and therefore discarded as well.
func DiscardEq(val int64, t Type) Instruction {
	return Instruction{op: OpDiscardEq, tp: t, val: val}
}

// DiscardEqBlock xx is a special type of discard so that a full record of data can be
// discarded. Since the size of the data to be discarded cannot be known in advance,
// the VM needs to consume it completely. The argument xx is the relative address of
// the routine that is used to consume the input.
func DiscardEqBlock(val int64, relPos int) Instruction {
	return Instruction{op: OpDiscardEq, tp: TypeBlock, pos: relPos, val: val}
}

// Skip avoids processing the current field, so it will remain zero-valued.
func Skip() Instruction {
	return Instruction{op: OpSkip}
}

// Jmp xx moves the current program counter to the relative position xx.
func Jmp(relPos int) Instruction {
	return Instruction{op: OpJmp, pos: relPos}
}

// JmpEq vv xx jumps to the relative position xx if the first int64 value of input
// matches vv. The int64 value is consumed this way and therefore discarded.
func JmpEq(val int64, relPos int) Instruction {
	return Instruction{op: OpJmpEq, pos: relPos, val: val}
}

// Call xx reads a record by calling the routine at relative position xx.
func Call(relPos int) Instruction {
	return Instruction{op: OpCall, pos: relPos}
}

// CallEq vv xx reads a record by calling the routine at relative position xx, if the
// first int64 value of input matches vv. The int64 value is consumed this way and
// therefore discarded.
func CallEq(val int64, relPos int) Instruction {
	return Instruction{op: OpCallEq, pos: relPos, val: val}
}

// Ret returns from a record reading routine.
func Ret() Instruction {
	return Instruction{op: OpRet}
}

// Loop xx reads as many blocks from input as it encounters, then jumps to the
// relative position xx.
func Loop(relPos int) Instruction {
	return Instruction{op: OpLoop, pos: relPos}
}

// LoopEq vv xx reads as many blocks from input as it encounters, then jumps to the
// relative position xx, if the first int64 value of input matches vv. The int64 value
// is consumed this way and therefore discarded.
func LoopEq(val int64, relPos int) Instruction {
	return Instruction{op: OpLoopEq, pos: relPos, val: val}
}

// EndLoop matches its corresponding Loop to signal the end of the block fields.
func EndLoop() Instruction {
	return Instruction{op: OpEndLoop}
}

// String is the implementation of Stringer for this instruction.
func (i Instruction) String() string {
	switch i.op {
	case OpError, OpHalt, OpLoad, OpSkip, OpRet, OpEndLoop:
		return i.op.String()

	case OpMov, OpDiscard:
		if i.op == OpDiscard && i.tp == TypeBlock {
			return fmt.Sprintf("%s %s\t--> %d", i.op, i.tp, i.pos)
		}
		return fmt.Sprintf("%s %s", i.op, i.tp)

	case OpMovEq, OpDiscardEq:
		if i.op == OpDiscardEq && i.tp == TypeBlock {
			return fmt.Sprintf("%s %d %s\t--> %d", i.op, i.val, i.tp, i.pos)
		}
		return fmt.Sprintf("%s %d %s", i.op, i.val, i.tp)

	case OpJmp, OpCall, OpLoop:
		return fmt.Sprintf("%s\t--> %d", i.op, i.pos)

	case OpJmpEq, OpCallEq, OpLoopEq:
		return fmt.Sprintf("%s %d\t--> %d", i.op, i.val, i.pos)

	case OpSort:
		return fmt.Sprintf("%s %v", i.op, i.val)

	default:
		return fmt.Sprintf("<invalid opCode %d>", i.op)
	}
}

// Size returns the serialized size of this instruction.
func (i Instruction) Size() int {
	switch i.op {
	case OpError, OpHalt, OpLoad, OpSkip, OpRet, OpEndLoop:
		return 1
	case OpMov, OpJmp, OpCall, OpLoop:
		return 2
	case OpDiscard:
		if i.tp != TypeBlock {
			return 2
		}
		return 3
	case OpMovEq, OpJmpEq, OpCallEq, OpLoopEq:
		return 3
	case OpDiscardEq:
		if i.tp != TypeBlock {
			return 3
		}
		return 4
	case OpSort:
		data := i.val.([]int)
		return 2 + len(data)
	default:
		panic(fmt.Sprintf("invalid instruction opCode %x", i.op))
	}
}

// WriteTo implements io.WriterTo for this instruction
func (i Instruction) WriteTo(w io.Writer) (n int64, err error) {
	buf := make([]byte, i.Size())
	buf[0] = byte(i.op)

	switch i.op {
	case OpError, OpHalt, OpLoad, OpSkip, OpRet, OpEndLoop:

	case OpMov:
		buf[1] = byte(i.tp)
	case OpJmp, OpCall, OpLoop:
		buf[1] = byte(i.pos)

	case OpDiscard:
		buf[1] = byte(i.tp)
		if i.tp == TypeBlock {
			buf[2] = byte(i.pos)
		}

	case OpMovEq:
		buf[1] = byte(i.val.(int64))
		buf[2] = byte(i.tp)
	case OpJmpEq, OpCallEq, OpLoopEq:
		buf[1] = byte(i.val.(int64))
		buf[2] = byte(i.pos)

	case OpDiscardEq:
		buf[1] = byte(i.val.(int64))
		buf[2] = byte(i.tp)
		if i.tp == TypeBlock {
			buf[3] = byte(i.pos)
		}

	case OpSort:
		data := i.val.([]int)
		l := len(data)
		buf[1] = byte(l)
		for i := 0; i < l; i++ {
			buf[2+i] = byte(data[i])
		}
	}
	var n2 int
	n2, err = w.Write(buf)
	return int64(n2), err
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
	case OpMovEq:
		inst = MovEq(int64(input[1]), Type(input[2]))
	case OpDiscard:
		t := Type(input[1])
		switch t {
		case TypeBlock:
			inst = DiscardBlock(relByteToInt(input[2]))
		default:
			inst = Discard(t)
		}
	case OpDiscardEq:
		t := Type(input[2])
		switch t {
		case TypeBlock:
			inst = DiscardEqBlock(int64(input[1]), relByteToInt(input[3]))
		default:
			inst = DiscardEq(int64(input[1]), t)
		}
	case OpSkip:
		inst = Skip()
	case OpJmp:
		inst = Jmp(relByteToInt(input[1]))
	case OpJmpEq:
		inst = JmpEq(int64(input[1]), relByteToInt(input[2]))
	case OpCall:
		inst = Call(relByteToInt(input[1]))
	case OpCallEq:
		inst = CallEq(int64(input[1]), relByteToInt(input[2]))
	case OpRet:
		inst = Ret()
	case OpLoop:
		inst = Loop(relByteToInt(input[1]))
	case OpLoopEq:
		inst = LoopEq(int64(input[1]), relByteToInt(input[2]))
	case OpEndLoop:
		inst = EndLoop()
	default:
		inst = Instruction{op: OpError}
	}
	return
}

func relByteToInt(b byte) int {
	if b < 128 {
		return int(b)
	}
	return -(int(^b) + 1)
}
