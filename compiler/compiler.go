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

func (c compiler) getOrCreateError(msg string) int {
	for i := range c.errors {
		if c.errors[i] == msg {
			return i
		}
	}
	c.errors = append(c.errors, msg)
	return len(c.errors) - 1
}

func (c compiler) compileType(m *method, wrt, rdr schema.GenericType) (err error) {
	// If the writer is not a union/optional type but the reader is, try and find
	// the first matching type in the reader as a target.
	if rdr != nil && !isUnionType(wrt) && isUnionType(rdr) {
		return c.compileNonUnionToUnion(m, wrt, rdr.(*schema.UnionType))
	}

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
		return c.compileUnion(m, wrtType, rdr)
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
		rdrType, ok := rdr.(*schema.ArrayType)
		if !ok && rdr != nil {
			return fmt.Errorf("incompatible types %s and %s", wrt.Name(), rdr.Name())
		}
		// Beware the nil interfaces!
		if rdr == nil {
			return c.compileBlock(m, wrtType, nil)
		}
		return c.compileBlock(m, wrtType, rdrType)
	case schema.GenericType:
		return c.compilePrimitive(m, wrt, rdr)
	}
	return fmt.Errorf("unsupported type %t", wrt)
}

func (c compiler) compilePrimitive(m *method, wrt, rdr schema.GenericType) error {
	if rdr != nil && !wrt.IsReadableBy(rdr, make(schema.VisitMap)) {
		return fmt.Errorf("incompatible types %s and %s", wrt.Name(), rdr.Name())
	}
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
	var rdrFields []schema.GenericType
	if rdr != nil {
		rdrFields = rdr.Children()
	}
	order, allAsc, err := getReadOrder(wrt.Children(), rdrFields, func(field *schema.FieldType) (rdrField *schema.FieldType, err error) {
		if rdr != nil {
			// All children within a record are fields, and must be found by name matching the given one
			if rdrField = rdr.FindFieldByNameOrAlias(field); rdrField != nil {
				if !field.Type().IsReadableBy(rdrField.Type(), make(schema.VisitMap)) {
					err = fmt.Errorf("incompatible schemas: field %s in reader has incompatible type in writer field %s", rdrField.Name(), field.Name())
				}
			}
		}
		return
	})
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

func (c compiler) compileUnion(m *method, wrt *schema.UnionType, rdr schema.GenericType) (err error) {
	var rdrFields []schema.GenericType
	if rdr != nil {
		if ct, ok := rdr.(schema.CompositeType); ok {
			rdrFields = ct.Children()
		} else {
			rdrFields = append(rdrFields, schema.NewField("non-union-field", rdr, 0))
		}
	}
	order, allAsc, err := getReadOrder(wrt.Children(), rdrFields, func(field *schema.FieldType) (rdrField *schema.FieldType, err error) {
		for i := range rdrFields {
			candidate := rdrFields[i].(*schema.FieldType)
			if field.Type().IsReadableBy(candidate.Type(), make(schema.VisitMap)) {
				if candidate.Type().Name() == field.Type().Name() {
					return candidate, nil
				}
				// The best option found up to date
				rdrField = candidate
			}
		}
		if rdrField == nil {
			err = fmt.Errorf("incompatible schemas: reader has no compatible fields for writer field %s in union", field.Name())
		}
		return
	})
	if err != nil {
		return err
	}

	if !allAsc {
		// Must invoke Sort for rearranging union fields
		refinedOrder := refineOrder(order)
		m.append(vm.Sort(refinedOrder))
	}

	m.append(vm.Load())
	skipJmpPos := -1
	if wrt.IsOptional() {
		// This jmp's rel offset has to be corrected later
		skipJmpPos = m.Size()
		m.append(vm.JmpEq(int64(wrt.OptionalIndex()), 0))
	}
	m.append(vm.Mov(vm.TypeAcc))

	sw := newSwitch()
	for i := range order {
		if i == wrt.OptionalIndex() && wrt.IsOptional() {
			continue
		}
		wrtField := asField(wrt.Children()[i])
		switch order[i] {
		case oDiscardable:
			fmt.Printf("Panic: discarding wrtField %d, that is a %s\n", wrtField.Index(), wrtField.Type().Name())
			panic("this should never happen in a union")
		case oSkippable:
			fmt.Printf("should this ever happen in a union?? I'm skipping wrtField %d, that is a %s\n", wrtField.Index(), wrtField.Type().Name())
			m.append(vm.Skip())
		default:
			typeMethod := newMethod("")
			var rType schema.GenericType
			if rdr != nil {
				f := asField(rdrFields[order[i]])
				rType = f.Type()
			}
			err = c.compileType(typeMethod, wrtField.Type(), rType)
			if err != nil {
				return
			}
			sw.addCase(int64(wrtField.Index()), typeMethod)
		}
	}
	sw.compileTo(m, c.getOrCreateError("invalid index for union"))
	if skipJmpPos > -1 {
		// Fix the optional jmp from the beginning
		jmpPos := m.Size() - skipJmpPos - 1
		m.code[skipJmpPos] = vm.JmpEq(int64(wrt.OptionalIndex()), jmpPos)
	}
	return nil
}

// compileNonUnionToUnion tries to find the first matching field in the reader
// for the giver writer type.
func (c compiler) compileNonUnionToUnion(m *method, wrt schema.GenericType, rdr *schema.UnionType) (err error) {
	for _, child := range rdr.Children() {
		childType := child.(*schema.FieldType).Type()
		if !wrt.IsReadableBy(childType, make(schema.VisitMap)) {
			continue
		}
		return c.compileType(m, wrt, childType)
	}
	return fmt.Errorf("incompatible types %s and %s", wrt.Name(), rdr.Name())
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
	case *schema.RecordType, *schema.MapType, *schema.ArrayType:
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
