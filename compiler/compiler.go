package compiler

import (
	"fmt"

	"github.com/actgardner/gogen-avro/schema"
	"github.com/actgardner/gogen-avro/vm"
)

func newCompiler() *compiler {
	return &compiler{
		methods: make(map[string]*method),
	}
}

type compiler struct {
	methods map[string]*method
	errors  []string
}

func (c compiler) compileType(m *method, wrt, rdr schema.GenericType) (err error) {
	assertNotReference(wrt, "writer")
	assertNotReference(rdr, "reader")

	/* TODO
	// If the writer is not a union but the reader is, try and find the first matching type in the union as a target
	if !isUnionType(wrt) && isUnionType(rdr) {
		return c.compileNonUnionToUnion(wrt, rdr.(*schema.UnionType))
	}
	*/

	switch wrtType := wrt.(type) {
	case *schema.RecordType:
		rdrType, ok := rdr.(*schema.RecordType)
		if !ok && rdr != nil {
			return fmt.Errorf("incompatible types %s and %s", wrt.Name(), rdr.Name())
		}
		recMethod, newlyCreated := c.getOrCreateMethodFor(wrtType, rdrType)
		if rdrType != nil {
			m.record(recMethod)
		} else {
			m.discardRecord(recMethod)
		}
		if newlyCreated {
			if err = c.compileRecord(recMethod, wrtType, rdrType); err == nil {
				recMethod.append(vm.Ret())
			}
		}
		return nil
	case *schema.FixedType:
		panic("fixed types are not implemented yet")
	case *schema.EnumType:
		panic("enum types are not implemented yet")
	case *schema.UnionType:
		panic("union types are not implemented yet")
	case *schema.MapType:
		rdrType, ok := rdr.(*schema.MapType)
		if !ok && rdr != nil {
			return fmt.Errorf("incompatible types %s and %s", wrt.Name(), rdr.Name())
		}
		// Beware the nil interfaces!
		if rdr == nil {
			return c.compileBlock(m, wrtType, nil)
		}
		return c.compileBlock(m, wrtType, rdrType)
	case *schema.ArrayType:
		panic("array types are not implemented yet")
	case schema.GenericType:
		return c.compilePrimitive(m, wrt, rdr)
	}
	return fmt.Errorf("unsupported type %t", wrt)
}

func (c compiler) compilePrimitive(m *method, wrt, rdr schema.GenericType) error {
	t, err := vmTypeFor(wrt)
	switch {
	case err != nil:
		return err
	case rdr == nil:
		m.append(vm.Discard(t))
	default:
		m.append(vm.Mov(t))
	}
	return nil
}

// compileBlock compiles a map or array type block
// TODO: create a special 'discard block' opType to discard block-type data
func (c compiler) compileBlock(m *method, wrt, rdr schema.GenericType) (err error) {
	discardMode := rdr == nil
	loopPos := m.block()

	// Maps are block types prefixed with a string key
	if _, ok := wrt.(*schema.MapType); ok {
		if !discardMode {
			m.append(vm.Mov(vm.TypeString))
		} else {
			m.append(vm.Discard(vm.TypeString))
		}
	}

	// All block types implement SingleChildType
	wrtChildType := wrt.(schema.SingleChildType).Type()
	var rdrChildType schema.GenericType
	if rdr != nil {
		rdrChildType = rdr.(schema.SingleChildType).Type()
	}
	if err = c.compileType(m, wrtChildType, rdrChildType); err != nil {
		return
	}

	m.append(vm.EndBlock())

	// Fix loop instruction
	pastEnd := len(m.code) - loopPos - 1
	if discardMode {
		m.code[loopPos] = vm.DiscardBlock(pastEnd)
	} else {
		m.code[loopPos] = vm.Block(pastEnd)
	}
	return nil
}

func (c compiler) compileRecord(m *method, wrt, rdr *schema.RecordType) (err error) {
	order, allAsc, err := getReadOrder(wrt, rdr)
	if err != nil {
		return err
	}
	if !allAsc {
		// Must invoke Sort for rearranging fields
		refinedOrder := refineOrder(order)
		m.append(vm.Sort(refinedOrder))
	}

	var wIdx, rIdx, maxRIdx int
	if rdr != nil {
		maxRIdx = len(rdr.Children())
	} else {
		maxRIdx = len(wrt.Children())
	}
	for i := range order {
		switch order[i] {
		case oDiscardable:
			f := asField(wrt.Children()[wIdx]).Type()
			c.discard(m, f)
			wIdx++
		case oSkippable:
			if rIdx < maxRIdx {
				m.append(vm.Skip())
				rIdx++
			}
		default:
			var rType schema.GenericType
			if rdr != nil {
				curIdx := order[i]
				f := asField(rdr.Children()[curIdx])
				if allAsc {
					for ; rIdx < curIdx; rIdx++ {
						m.append(vm.Skip())
					}
				}
				rType = f.Type()
			}
			err = c.compileType(m, asField(wrt.Children()[wIdx]).Type(), rType)
			if err != nil {
				return
			}
			wIdx++
			rIdx++
		}
	}
	return nil
}

// getOrCreateMethodFor returns an existing method for the given type,
// or creates a new one, if not exists. The returning bool value will be
// true if the method has been created anew.
func (c *compiler) getOrCreateMethodFor(w, r schema.GenericType) (*method, bool) {
	var name string
	if r == nil {
		name = fmt.Sprintf("r-%s", w.Name())
	} else {
		name = fmt.Sprintf("rw-%s", w.Name())
	}

	typeMethod, ok := c.methods[name]
	if !ok {
		typeMethod = newMethod(name)
		c.methods[name] = typeMethod
	}
	return typeMethod, !ok
}

func (c compiler) discard(m *method, t schema.GenericType) error {
	switch t.(type) {
	case *schema.RecordType, *schema.MapType:
		return c.compileType(m, t, nil)
	default:
		ft, err := vmTypeFor(t)
		if err != nil {
			return err
		}
		m.append(vm.Discard(ft))
	}
	return nil
}
