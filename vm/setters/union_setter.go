package setters

import (
	"fmt"
	"reflect"
)

// BaseUnion is the base type for all unions
type BaseUnion struct {
	Type  int64
	Value interface{}
}

// Self-locator for BaseUnion
func (u *BaseUnion) self() *BaseUnion {
	return u
}

type unionHelper interface {
	self() *BaseUnion
	UnionTypes() []reflect.Type
}

func newUnionSetter(union interface{}) *unionSetter {
	helper := union.(unionHelper)
	types := helper.UnionTypes()
	l := len(types)
	fields := make([]interface{}, l, l)
	for i := range fields {
		fields[i] = types[i]
	}
	return &unionSetter{
		sortableFieldsComponent: newSortableFieldsComponent(fields),
		baseUnion:               helper.self(),
	}
}

type unionSetter struct {
	exhaustNotifierComponent
	sortableFieldsComponent
	currentField int
	baseUnion    *BaseUnion
}

func (s *unionSetter) reset() (err error) {
	s.currentField = 0
	s.initSortOrder()
	return nil
}

// Union setter can be initialized with the order its types
// are to be assigned.
func (s *unionSetter) Init(arg interface{}) error {
	positions, ok := arg.([]int)
	if ok {
		s.sort(positions)
		return nil
	}
	return fmt.Errorf("struct setter initialization expects []int, got %T", positions)
}

// Field 0 is the union type discriminator, 1 is the value
func (s *unionSetter) IsExhausted() bool {
	return s.currentField > 1
}

// GetInner can only be called for complex value types, and after the type disciminator
// has been read.
func (s *unionSetter) GetInner() (inner Setter, err error) {
	if s.currentField != 1 {
		return nil, ErrTypeNotSupported
	}

	// Create the appropriate union type element
	t := s.get(int(s.baseUnion.Type)).(reflect.Type)
	valueElem := reflect.New(t.Elem())

	// Try and match it as a complex type
	if inner, err = NewSetterFor(valueElem.Elem().Addr().Interface()); err == nil {
		// Must install a notification cb for setting the copy onto the interface field
		inner.setExhaustCallback(func(_ Setter) {
			reflect.ValueOf(&s.baseUnion.Value).Elem().Set(valueElem.Elem().Convert(t.Elem()))
			s.goNext()
		})
	}
	return
}

// Execute performs the given operation and advances its current field pointer,
// if apply. If the current field happens to be a nested field, the operation is applied
// recursively. In this case, the current pointer cannot be advanced until the inner
// setter gets exhausted. Executing past the last field returns ErrSetterEOF.
func (s *unionSetter) Execute(op OperationType, value interface{}) error {
	switch s.currentField {
	case 0:
		// Only SetField makes sense to execute. SkipField always succeeds.
		if op != SetField {
			break
		}
		s.baseUnion.Type = value.(int64)
	case 1:
		if inner, err := s.GetInner(); err == nil {
			// If there is a nested Setter, use it for this field. Nested triggers goNext when exhausted.
			return inner.Execute(op, value)
		}

		// Only SetField makes sense to execute. SkipField always succeeds.
		if op != SetField {
			break
		}
		s.baseUnion.Value = value
	}

	s.goNext()
	return nil
}

// goNext advances internal current pointer. No error checking is performed.
func (s *unionSetter) goNext() {
	s.currentField++
	if s.IsExhausted() && s.hasExhaustCallback() {
		s.trigger(s)
	}
}
