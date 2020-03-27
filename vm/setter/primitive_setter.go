package setter

import (
	"errors"
	"fmt"
	"reflect"
)

type primitiveSetter struct {
	curFld int
	fields []interface{}
	nested Setter
}

// Init is always referred to leaf node of the graph, so redirect it, if able.
// Primitive setters need no initialization, otherwise.
func (s *primitiveSetter) Init(arg interface{}) error {
	if s.nested == nil {
		s.initNested()
		if s.nested == nil {
			return ErrNoInitialization
		}
	}
	return s.nested.Init(arg)
}

// Set puts the value into the current field this Setter is tracking.
// A successful set advances the current field. Setting past the last
// field returns ErrSetterEOF.
func (s *primitiveSetter) Set(value interface{}) error {
	ok, err := s.setNested(value)
	if ok || err != nil {
		return err
	}

	if err = s.setCurrent(value); err == nil {
		// Advance only if not currently set a nested field
		if s.nested == nil {
			s.curFld++
			// Must signal EOF asap, otherwise map/array setters could leave
			// unassigned data
			err = s.checkEOF()
		}
	}
	return err
}

// Reset restarts the setter to its initial values, so their fields can be set again.
func (s *primitiveSetter) Reset() error {
	s.curFld = 0
	s.nested = nil
	return nil
}

// setNested tries to set the value using a nested setter, so that
// complex, non-primitive types can be traversed.
func (s *primitiveSetter) setNested(value interface{}) (bool, error) {
	// No nested, try to match current field
	if s.nested == nil {
		s.initNested()
		// Still no luck, current field is not a setter, go on.
		if s.nested == nil {
			return false, nil
		}
	}

	// Successful set in nested, return success
	err := s.nested.Set(value)
	if err == nil {
		return true, nil
	}

	// Unsuccessful set in nested not caused by EOF in nested setter,
	// return error status
	if !errors.Is(err, ErrSetterEOF) {
		return false, err
	}

	// Nested setter fully exhausted: disable its use and advance
	// pointer for using the following field in this setter.
	s.nested = nil
	s.curFld++
	return true, s.checkEOF()
}

// initNested tries to initialize a nested setter from current field
func (s *primitiveSetter) initNested() error {
	fld, err := s.getCurrentField()
	if err != nil {
		return err
	}
	s.nested = s.asSetter(fld)
	return nil
}

// asSetter tries to get a Setter from given fld by checking if the
// field is already a Setter. If so, then uses it directly; otherwise,
// if the field implemets setter.Factory, a setter is created using it.
// If all the above fails, nil is returned.
func (s *primitiveSetter) asSetter(fld interface{}) Setter {
	// Check if field is already a Setter
	if setter, ok := fld.(Setter); ok {
		return setter
	}

	// Check if it is a Factory instead
	if factory, ok := fld.(Factory); ok {
		return factory.Setter()
	}

	// Well, let's check for an autodetectable type
	if setter, err := NewSetterFor(fld); err == nil {
		return setter
	}

	// Omit error, since this field cannot be returned as Setter
	return nil
}

// getCurrentField gets the current field to be set.
// Overflow condition is checked and returned as error.
func (s *primitiveSetter) getCurrentField() (fld interface{}, err error) {
	if err = s.checkEOF(); err == nil {
		fld = s.fields[s.curFld]
	}
	return
}

func (s *primitiveSetter) checkEOF() (err error) {
	if s.curFld >= len(s.fields) {
		err = ErrSetterEOF
	}
	return
}

// setCurrent sets the value into the current field.
func (s *primitiveSetter) setCurrent(value interface{}) error {
	fld, err := s.getCurrentField()
	if err != nil {
		return err
	}
	fldValue := reflect.ValueOf(fld)
	if fldValue.Kind() == reflect.Ptr {
		return s.setPointerElem(fldValue.Elem(), value)
	}
	return fmt.Errorf("type %s is not supported", fldValue.Kind())
}

// setPointerElem sets the value into the current field, which must be a pointer.
func (s *primitiveSetter) setPointerElem(elem reflect.Value, value interface{}) error {
	// If the target elem is an indirection, its pointer must be created first
	if elem.Type().Kind() == reflect.Ptr {
		indirectElem := reflect.New(elem.Type().Elem())
		elem.Set(indirectElem)

		// Try nesting: if it happens that the target elem type is
		// a Setter/Factory, use it instead
		if indirectElem.CanInterface() {
			if s.nested = s.asSetter(indirectElem.Interface()); s.nested != nil {
				return s.nested.Set(value)
			}
		}

		// Otherwise, fallthrough the target item type
		elem = indirectElem.Elem()
	}

	// elem.Set(reflect.Zero(elem.Type()))
	elem.Set(reflect.ValueOf(value))
	return nil
}
