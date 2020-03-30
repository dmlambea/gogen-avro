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

	callFn := func(s setters.Setter) {
		var innerSetter setters.Setter
		if innerSetter, err = s.GetInner(); err != nil {
			return
		}
		err = e.doRun(depth+1, pc+inst.pos+1, innerSetter)
	}

	loopFn := func(s setters.Setter) {
		var innerSetter setters.Setter
		if innerSetter, err = s.GetInner(); err != nil {
			return
		}
		err = e.runLoop(depth+1, pc+1, innerSetter)
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
		case OpDiscard:
			if inst.tp != TypeBlock {
				_, err = readInput(inst.tp, e.input)
			} else {
				err = e.doRun(depth+1, inst.pos, setters.NewSkipperSetter())
			}
		case OpDiscardEq:
			var fn func()
			if inst.tp != TypeBlock {
				fn = func() {
					_, err = readInput(inst.tp, e.input)
				}
			} else {
				fn = func() {
					err = e.doRun(depth+1, inst.pos, setters.NewSkipperSetter())
				}
			}
			if loadFn(); err == nil {
				if acc == int64(inst.val.(int32)) {
					fn()
				}
			}
		case OpSkip:
			err = setter.Execute(setters.SkipField, nil)
		case OpJmp:
			pc += inst.pos
		case OpJmpEq:
			if acc == int64(inst.val.(int)) {
				pc += inst.pos
			}
		case OpCall:
			callFn(setter)
		case OpCallEq:
			eqFn(func() {
				callFn(setter)
			})
		case OpLoop:
			loopFn(setter)
			pc += inst.pos
		case OpLoopEq:
			eqFn(func() {
				loopFn(setter)
			})
			pc += inst.pos
		case OpRet, OpEndLoop:
			if depth == 0 {
				err = fmt.Errorf("can't %s from main flow at %d", inst.op, pc)
			}
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
