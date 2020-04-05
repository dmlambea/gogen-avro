package setters

import (
	"reflect"
)

func newUnionSetter(union interface{}) *unionSetter {
	return &unionSetter{helper: union.(unionHelper)}
}

type unionSetter struct {
	exhaustNotifierComponent
	curFld int
	helper unionHelper
	inner  Setter
}

func (s *unionSetter) reset() (err error) {
	s.curFld = 0
	return nil
}

// Primitive setters allow for no initialization.
// TODO use this to infor about the order of fields.
func (s *unionSetter) Init(arg interface{}) error {
	return ErrNoInitialization
}

// Field 0 is the union type discriminator, 1 is the value
func (s *unionSetter) IsExhausted() bool {
	return s.curFld > 1
}

// GetInner can only be called for complex value types, and after the type disciminator
// has been read.
func (s *unionSetter) GetInner() (inner Setter, err error) {
	if s.curFld != 1 {
		return nil, ErrTypeNotSupported
	}

	if s.inner == nil {
		return nil, ErrTypeNotSupported
	}
	return s.inner, nil
}

// Execute performs the given operation and advances its current field pointer,
// if apply. If the current field happens to be a nested field, the operation is applied
// recursively. In this case, the current pointer cannot be advanced until the inner
// setter gets exhausted. Executing past the last field returns ErrSetterEOF.
func (s *unionSetter) Execute(op OperationType, value interface{}) (err error) {
	base := s.helper.self()
	switch s.curFld {
	case 0:
		// Only SetField makes sense to execute. SkipField always succeeds.
		if op != SetField {
			break
		}
		base.Type = value.(int64)

		// Create the appropriate union type element
		valueElem := reflect.New(s.helper.UnionTypes()[base.Type].Elem())

		// Try and match it as a complex type
		if s.inner, err = NewSetterFor(valueElem.Elem().Addr().Interface()); err == nil {
			// Must install a notification cb for setting the copy onto the interface field
			s.inner.setExhaustCallback(func(_ Setter) {
				reflect.ValueOf(&base.Value).Elem().Set(valueElem.Elem())
				s.goNext()
			})
		}
		err = nil // Disable nested setter's error propagation
	case 1:
		// If there is a nested Setter, use it for this field. Nested triggers goNext when exhausted.
		if s.inner != nil {
			return s.inner.Execute(op, value)
		}

		// Only SetField makes sense to execute. SkipField always succeeds.
		if op != SetField {
			break
		}
		base.Value = value
	}

	if err == nil {
		s.goNext()
	}
	return
}

// goNext advances internal current pointer. No error checking is performed.
func (s *unionSetter) goNext() {
	s.curFld++
	if s.IsExhausted() && s.hasExhaustCallback() {
		s.trigger(s)
	}
}
