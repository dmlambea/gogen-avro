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
	var obj interface{}

	for {
		inst := e.prog.instructions[pc]
		switch inst.op {
		case OpError:
			return fmt.Errorf("bad instruction %+v at %d", inst, pc)
		case OpHalt:
			return nil
		case OpSort:
			if err = setter.Init(inst.val.([]int)); err != nil {
				return err
			}
		case OpLoad:
			if acc, err = readLong(e.input); err != nil {
				return err
			}
		case OpMov:
			if obj, err = readInput(inst.tp, e.input); err != nil {
				return err
			}
			if err = setter.Execute(setters.SetField, obj); err != nil {
				return err
			}
		case OpMovOpt:
			if acc, err = readLong(e.input); err != nil {
				return err
			}
			switch {
			case acc == int64(inst.val.(int)):
				if obj, err = readInput(inst.tp, e.input); err != nil {
					return err
				}
				if err = setter.Execute(setters.SetField, obj); err != nil {
					return err
				}
			default:
				if err = setter.Execute(setters.SkipField, nil); err != nil {
					return err
				}
			}
		case OpDiscard:
			if obj, err = readInput(inst.tp, e.input); err != nil {
				return err
			}
		case OpSkip:
			if err = setter.Execute(setters.SkipField, nil); err != nil {
				return err
			}
		case OpJmp:
			pc = inst.pos
			// Avoid incrementing the PC
			continue
		case OpJmpEq:
			if acc == int64(inst.val.(int)) {
				pc = inst.pos
				// Avoid incrementing the PC if the jump succeeds
				continue
			}
		case OpCall:
			var innerSetter setters.Setter
			if innerSetter, err = setter.GetInner(); err != nil {
				return err
			}
			if err = e.doRun(depth+1, inst.pos, innerSetter); err != nil {
				return err
			}
		case OpLoopStart:
			var innerSetter setters.Setter
			if innerSetter, err = setter.GetInner(); err != nil {
				return err
			}
			if err = e.runLoop(depth+1, pc+1, innerSetter); err != nil {
				return
			}
			pc = inst.pos
			// Avoid incrementing the PC
			continue
		case OpRet, OpLoopEnd:
			if depth == 0 {
				return fmt.Errorf("can't %s from main flow at %d", inst.op, pc)
			}
			return
		}
		pc++
	}
	return nil
}

// runLoop is a convenience method to allow running over block-serialized types (maps, arrays)
func (e engine) runLoop(depth, pc int, setter setters.Setter) (err error) {
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
