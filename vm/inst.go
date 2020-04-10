package vm

import (
	"bytes"
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

// Ret returns from a record reading routine.
func Ret() Instruction {
	return Instruction{op: OpRet}
}

// Halt stops the VM.
// TODO: add exit code for erroring.
func Halt(val int64) Instruction {
	return Instruction{op: OpHalt, val: val}
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

// MovFixed loads "amount" data of type "fixed" into current field.
func MovFixed(amount int64) Instruction {
	return Instruction{op: OpMov, tp: TypeFixed, val: amount}
}

// Discard tt discards input data of type tt.
func Discard(t Type) Instruction {
	return Instruction{op: OpDiscard, tp: t}
}

// DiscardFixed discards "amount" bytes of data.
func DiscardFixed(amount int64) Instruction {
	return Instruction{op: OpDiscard, tp: TypeFixed, val: amount}
}

// DiscardRecord xx is a special type of discard so that a full record of data can be
// discarded. Since the size of the data to be discarded cannot be known in advance,
// the VM needs to consume it completely. The argument xx is the relative address of
// the code that is used to consume the input.
func DiscardRecord(relPos int) Instruction {
	return Instruction{op: OpDiscard, tp: TypeRecord, pos: relPos}
}

// DiscardBlock xx is a special type of discard so that a full stream of data blocks can be
// discarded. Since the size of the data to be discarded cannot be always known in advance,
// the VM needs to consume it completely. The argument xx is the relative address of
// the code that is used to consume the input.
func DiscardBlock(relPos int) Instruction {
	return Instruction{op: OpDiscard, tp: TypeBlock, pos: relPos}
}

// Skip avoids processing the current field, so it will remain zero-valued.
func Skip() Instruction {
	return Instruction{op: OpSkip}
}

// Jmp xx moves the current program counter to the relative position xx.
func Jmp(relPos int) Instruction {
	return Instruction{op: OpJmp, pos: relPos}
}

// Case vv xx jumps to the relative position xx if the first int64 value of input
// matches vv. The int64 value is consumed this way and therefore discarded.
func Case(val int64, relPos int) Instruction {
	return Instruction{op: OpCase, pos: relPos, val: val}
}

// SkipCase vv xx skips the current field and jumps to the relative position xx
// if the first int64 value of input matches vv. The int64 value is consumed this
// way and therefore discarded.
func SkipCase(val int64, relPos int) Instruction {
	return Instruction{op: OpSkipCase, pos: relPos, val: val}
}

// Record xx reads a record by calling the routine at relative position xx.
func Record(relPos int) Instruction {
	return Instruction{op: OpRecord, pos: relPos}
}

// Block xx reads as many blocks from input as it encounters, then jumps to the
// relative position xx.
func Block(relPos int) Instruction {
	return Instruction{op: OpBlock, pos: relPos}
}

// EndBlock matches its corresponding Block to signal the end of the block fields.
func EndBlock() Instruction {
	return Instruction{op: OpEndBlock}
}

// Opcode returns the opcode for this instruction
func (i Instruction) Opcode() Opcode {
	return i.op
}

// SetPos sets the relative position of this jump-type instruction.
// This will panic if this is not a jump-type instruction.
func (i *Instruction) SetPos(pos int) {
	if !i.IsJumpType() {
		panic(fmt.Sprintf("%s is not a jump-type instruction", i))
	}
	i.pos = pos
}

// IsRecordType returns true if this instruction can call a subroutine
// for either reading or discarding a record.
func (i Instruction) IsRecordType() bool {
	switch i.op {
	case OpRecord:
		return true
	case OpDiscard:
		return i.tp == TypeRecord
	default:
		return false
	}
}

// IsJumpType returns true if this instruction can make the VM to move its program counter
// to a relative position counting after the next instruction in the program.
func (i Instruction) IsJumpType() bool {
	switch i.op {
	case OpJmp, OpRecord, OpBlock, OpCase, OpSkipCase:
		return true
	case OpDiscard:
		return i.tp == TypeBlock || i.tp == TypeRecord
	default:
		return false
	}
}

// IsBlockType returns true if this instruction can make the VM to consume/discard block-encoded data.
func (i Instruction) IsBlockType() bool {
	switch i.op {
	case OpBlock:
		return true
	case OpDiscard:
		return i.tp == TypeBlock
	default:
		return false
	}
}

// String is the implementation of Stringer for this instruction.
func (i Instruction) String() string {
	switch i.op {
	case OpError, OpRet, OpLoad, OpSkip, OpEndBlock:
		return i.op.String()

	case OpHalt:
		return fmt.Sprintf("%s (%d)", i.op, i.val)

	case OpMov, OpDiscard:
		switch {
		case i.IsJumpType():
			// For discard block/record types
			return fmt.Sprintf("%s %s\t--> %d", i.op, i.tp, i.pos)
		case i.tp == TypeFixed:
			return fmt.Sprintf("%s %s [%d]", i.op, i.tp, i.val)
		default:
			return fmt.Sprintf("%s %s", i.op, i.tp)
		}

	case OpJmp, OpRecord, OpBlock:
		return fmt.Sprintf("%s\t--> %d", i.op, i.pos)

	case OpCase, OpSkipCase:
		return fmt.Sprintf("%s %d\t--> %d", i.op, i.val, i.pos)

	case OpSort:
		return fmt.Sprintf("%s %v", i.op, i.val)

	default:
		return fmt.Sprintf("<invalid opCode %d>", i.op)
	}
}

// readFrom is somehow similar to io.ReaderFrom for this instruction, so it reads the
// serialized form of this instruction from the reader r, but no total bytes read count
// is returned.
func (i *Instruction) readFrom(r io.Reader) (err error) {
	// Opcode
	if b, err := readByte(r); err != nil {
		return err
	} else {
		i.op = Opcode(b)
	}

	asType := func(val interface{}, err error) (Type, error) {
		if err != nil {
			return TypeError, err
		}
		return Type(val.(byte)), nil
	}

	asInt := func(val interface{}, err error) (int, error) {
		if err != nil {
			return 0, err
		}
		if i64, ok := val.(int64); ok {
			return int(i64), nil
		}
		return int(val.(int32)), nil
	}

	switch i.op {
	case OpError, OpLoad, OpSkip, OpRet, OpEndBlock:

	case OpHalt:
		if i.val, err = readLong(r); err != nil {
			return
		}

	case OpMov:
		if i.tp, err = asType(readByte(r)); err != nil {
			return
		}
		if i.tp == TypeFixed {
			if i.val, err = readLong(r); err != nil {
				return
			}
		}

	case OpJmp, OpRecord, OpBlock:
		if i.pos, err = asInt(readLong(r)); err != nil {
			return
		}

	case OpDiscard:
		if i.tp, err = asType(readByte(r)); err != nil {
			return
		}
		if i.IsJumpType() {
			if i.pos, err = asInt(readLong(r)); err != nil {
				return
			}
		} else if i.tp == TypeFixed {
			if i.val, err = readLong(r); err != nil {
				return
			}
		}

	case OpCase, OpSkipCase:
		if i.val, err = readLong(r); err != nil {
			return
		}
		if i.pos, err = asInt(readLong(r)); err != nil {
			return
		}

	case OpSort:
		var j, l int32
		if l, err = readInt(r); err != nil {
			return
		}
		data := make([]int, l)
		for j = 0; j < l; j++ {
			if data[j], err = asInt(readInt(r)); err != nil {
				return
			}
		}
		i.op = OpSort
		i.val = data

	default:
		i.op = OpError
	}
	return
}

// WriteTo implements io.WriterTo for this instruction and writes the
// serialized form of this instruction on the writer w.
func (i Instruction) WriteTo(w io.Writer) (n int64, err error) {
	var buf bytes.Buffer

	// Opcode
	WriteByte(byte(i.op), &buf)

	// Operands, depending on the instruction kind
	switch i.op {
	case OpError, OpLoad, OpSkip, OpRet, OpEndBlock:

	case OpHalt:
		WriteLong(i.val.(int64), &buf)

	case OpMov:
		WriteByte(byte(i.tp), &buf)
		if i.tp == TypeFixed {
			WriteLong(i.val.(int64), &buf)
		}

	case OpJmp, OpRecord, OpBlock:
		WriteLong(int64(i.pos), &buf)

	case OpDiscard:
		WriteByte(byte(i.tp), &buf)
		if i.IsJumpType() {
			WriteLong(int64(i.pos), &buf)
		} else if i.tp == TypeFixed {
			WriteLong(i.val.(int64), &buf)
		}

	case OpCase, OpSkipCase:
		WriteLong(i.val.(int64), &buf)
		WriteLong(int64(i.pos), &buf)

	case OpSort:
		data := i.val.([]int)
		l := len(data)
		WriteInt(int32(l), &buf)
		for i := 0; i < l; i++ {
			WriteInt(int32(data[i]), &buf)
		}
	}

	var n2 int
	n2, err = w.Write(buf.Bytes())
	return int64(n2), err
}
