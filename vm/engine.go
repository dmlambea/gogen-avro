package vm

import (
	"fmt"

	"github.com/actgardner/gogen-avro/vm/setters"
)

type engine struct {
	prog       Program
	input      ByteReader
	mainSetter setters.Setter
}

func NewEngine(p Program, setter setters.Setter) engine {
	return engine{
		prog:       p,
		mainSetter: setter,
	}
}

// Run starts the engine
func (e engine) Run(input ByteReader) error {
	e.input = input
	return e.doRun(0, 0, e.mainSetter)
}

func (e engine) doRun(depth, pc int, setter setters.Setter) (err error) {
	var acc int64
	var inst Instruction

	loadFn := func() {
		acc, err = readLong(e.input)
	}

	movFn := func() {
		var obj interface{}
		obj, err = readInput(inst.tp, e.input)
		if err == nil {
			err = setter.Execute(setters.SetField, obj)
		}
	}

	recordFn := func() {
		var innerSetter setters.Setter
		if innerSetter, err = setter.GetInner(); err != nil {
			return
		}
		err = e.doRun(depth+1, pc+inst.pos+1, innerSetter)
	}

	discardFn := func() {
		_, err = readInput(inst.tp, e.input)
	}

	discardBlockFn := func() {
		if err = e.runBlock(depth+1, pc+1, setters.NewSkipperSetter()); err == nil {
			pc += inst.pos
		}
	}

	discardRecordFn := func() {
		err = e.doRun(depth+1, pc+inst.pos+1, setters.NewSkipperSetter())
	}

	blockFn := func() {
		var innerSetter setters.Setter
		if innerSetter, err = setter.GetInner(); err != nil {
			return
		}
		if err = e.runBlock(depth+1, pc+1, innerSetter); err == nil {
			pc += inst.pos
		}
	}

	eqFn := func(fn func()) {
		if loadFn(); err == nil {
			if acc == inst.val.(int64) {
				fn()
			} else {
				err = setter.Execute(setters.SkipField, nil)
			}
		}
	}

	for {
		inst = e.prog.instructions[pc]
		switch inst.op {
		case OpError:
			err = fmt.Errorf("bad instruction %+v at %d", inst, pc)
		case OpHalt:
			return nil
		case OpSort:
			err = setter.Init(inst.val.([]int))
		case OpLoad:
			loadFn()
		case OpMov:
			movFn()
		case OpMovEq:
			eqFn(movFn)
		case OpDiscardEq:
			// No matter the value of acc, the reader's field must be discarded
			loadFn()
			fallthrough
		case OpDiscard:
			switch {
			case inst.IsJumpType() == false:
				discardFn()
			case inst.IsBlockType():
				discardBlockFn()
			default:
				discardRecordFn()
			}
		case OpSkip:
			err = setter.Execute(setters.SkipField, nil)
		case OpJmp:
			pc += inst.pos
		case OpJmpEq:
			if acc == int64(inst.val.(int)) {
				pc += inst.pos
			}
		case OpRecord:
			recordFn()
		case OpRecordEq:
			eqFn(recordFn)
		case OpBlock:
			blockFn()
		case OpBlockEq:
			eqFn(blockFn)
		case OpRet, OpEndBlock:
			return
		}
		// General error occurred in executing the instruction
		if err != nil {
			return
		}
		pc++
	}
	return
}

// runBlock is a convenience method to allow running over block-serialized types (maps, arrays)
func (e engine) runBlock(depth, pc int, setter setters.Setter) (err error) {
	for {
		// Load block length. If no more blocks (lenght==0) or an error occurs, go back
		count, err := readLong(e.input)
		if count == 0 || err != nil {
			return err
		}

		// If input is signalling a blocksize indicator
		if count < 0 {
			// Ignore blocksize indicator: use abs(count) as count instead
			if _, err = readLong(e.input); err != nil {
				return err
			}
			count = -count
		}

		// Inform the block owner's setter about the number of items to be expected within this block
		setter.Init(int(count))

		innerSetter, err := setter.GetInner()
		if err != nil {
			return err
		}
		for ; count > 0; count-- {
			// Consume one item type each time
			if err = e.doRun(depth, pc, innerSetter); err != nil {
				return err
			}
		}
	}
}

func readInput(dataType Type, input ByteReader) (obj interface{}, err error) {
	switch dataType {
	case TypeNull:
		return nil, nil
	case TypeBool:
		return readBool(input)
	case TypeInt:
		return readInt(input)
	case TypeLong:
		return readLong(input)
	case TypeFloat:
		return readFloat(input)
	case TypeDouble:
		return readDouble(input)
	case TypeString:
		return readString(input)
	case TypeBytes:
		return readBytes(input)
	}
	return nil, fmt.Errorf("bad data type %x", dataType)
}
