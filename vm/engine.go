package vm

import (
	"fmt"
	"io"

	"github.com/actgardner/gogen-avro/vm/setters"
)

// Engine is the structure representing a program runner.
type Engine struct {
	Program     Program
	StackTraces bool
	input       io.Reader
}

// Run executes this engine's program agains the given input. Target must be the
// address of a variable able to hold the produced output.
func (e Engine) Run(input io.Reader, targetAddr interface{}) error {
	targetSetter, err := setters.NewSetterFor(targetAddr)
	if err != nil {
		return fmt.Errorf("unable to create a setter for target: %w", err)
	}
	return e.RunWithSetter(input, targetSetter)
}

// RunWithSetter executes this engine's program agains the given input and producing
// its output via the given setter.
func (e Engine) RunWithSetter(input io.Reader, setter setters.Setter) error {
	e.input = input
	return e.runMethod(0, setter)
}

// runMethod executes code at the given program counter, using the given setter.
func (e Engine) runMethod(pc int, setter setters.Setter) (err error) {
	var acc int64
	var inst Instruction

	for {
		inst = e.Program.instructions[pc]
		switch inst.op {
		case OpError:
			err = fmt.Errorf("bad instruction %+v at %d", inst, pc)
		case OpHalt:
			err = fmt.Errorf("execution halted: %s", e.Program.errors[int(inst.val.(int64))])
		case OpSort:
			err = setter.Init(inst.val.([]int))
		case OpLoad:
			acc, err = readLong(e.input)
		case OpMov:
			var obj interface{}
			switch inst.tp {
			case TypeAcc:
				obj = acc
			default:
				obj, err = readInput(inst, e.input)
			}
			if err == nil {
				err = setter.Execute(setters.SetField, obj)
			}
		case OpDiscard:
			switch {
			case inst.IsJumpType() == false:
				_, err = readInput(inst, e.input)
			case inst.IsBlockType():
				if err = e.runBlock(pc+1, setters.NewSkipperSetter()); err == nil {
					pc += inst.pos
				}
			default:
				err = e.runMethod(pc+inst.pos+1, setters.NewSkipperSetter())
			}
		case OpSkip:
			err = setter.Execute(setters.SkipField, nil)
		case OpJmp:
			pc += inst.pos
		case OpCase:
			if acc == inst.val.(int64) {
				pc += inst.pos
			}
		case OpSkipCase:
			if acc == inst.val.(int64) {
				err = setter.Execute(setters.SkipField, nil)
				pc += inst.pos
			}
		case OpRecord:
			var innerSetter setters.Setter
			if innerSetter, err = setter.GetInner(); err == nil {
				err = e.runMethod(pc+inst.pos+1, innerSetter)
			}
		case OpBlock:
			var innerSetter setters.Setter
			if innerSetter, err = setter.GetInner(); err == nil {
				if err = e.runBlock(pc+1, innerSetter); err == nil {
					pc += inst.pos
				}
			}
		case OpRet, OpEndBlock:
			return
		}
		// General error occurred in executing the instruction
		if err != nil {
			if e.StackTraces {
				err = fmt.Errorf("%s\n at pc %d: '%s'", err.Error(), pc, inst)
			}
			return
		}
		pc++
	}
	return
}

// runBlock is a convenience method to allow running code dealing with
// block-serialized types (maps, arrays), since they require a loop.
func (e Engine) runBlock(pc int, setter setters.Setter) (err error) {
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
			if err = e.runMethod(pc, innerSetter); err != nil {
				return err
			}
		}
	}
}

// readInput is a convenience function for calling the appropriate reader, depending
// on the data type of the instruction invoking the read. Type 'acc' is not covered
// here, because the accumulator is not read from the program input.
func readInput(inst Instruction, input io.Reader) (obj interface{}, err error) {
	switch inst.tp {
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
	case TypeFixed:
		return readFixed(input, int(inst.val.(int64)))
	}
	return nil, fmt.Errorf("unable to read data of type %s from input", inst.tp)
}
