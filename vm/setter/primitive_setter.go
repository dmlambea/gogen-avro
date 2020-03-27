package setter

import (
	"errors"
	"fmt"
	"reflect"
)

type operation byte

const (
	opSet  operation = iota // The requested operation is to set the current field
	opSkip                  // The requested operation is to skip the current field
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

// Set puts the value into the current field this Setter is tracking. If the current
// field is a nested Setter, the Set operation is performed recursively until
// success. Setting the last non-nested field returns ErrSetterEOF.
func (s *primitiveSetter) Set(value interface{}) error {
	return s.doOperation(opSet, value)
}

// Skip advances the pointer of current field this Setter is tracking. If the current
// field is a nested Setter, the Skip operation is performed recursively until
// success. Skipping the last non-nested field returns ErrSetterEOF.
func (s *primitiveSetter) Skip() error {
	return s.doOperation(opSkip, nil)
}

// Reset restarts the setter to its initial values, so their fields can be set again.
func (s *primitiveSetter) Reset() error {
	s.curFld = 0
	s.nested = nil
	return nil
}

// doOperation executes the given operation
func (s *primitiveSetter) doOperation(op operation, value interface{}) error {
	ok, err := s.doNested(op, value)
	if ok || err != nil {
		return err
	}

	if err = s.doCurrent(op, value); err == nil {
		// Advance only if not currently operating on a nested field
		if s.nested == nil {
			s.curFld++
			// Must signal EOF asap, otherwise map/array setters could leave
			// unassigned/unskipped data
			err = s.checkEOF()
		}
	}
	return err
}

// doNested tries to perform the operation using a nested setter, so that
// complex, non-primitive types can be traversed.
func (s *primitiveSetter) doNested(op operation, value interface{}) (bool, error) {
	// No nested, try to match current field
	if s.nested == nil {
		s.initNested()
		// Still no luck, current field is not a setter, go on.
		if s.nested == nil {
			return false, nil
		}
	}

	// If successfully operated in nested, return success.
	var err error
	switch op {
	case opSet:
		err = s.nested.Set(value)
	case opSkip:
		err = s.nested.Skip()
	}
	if err == nil {
		return true, nil
	}

	// Unsuccessful operation in nested, not caused
	// by EOF in nested setter -> return error status
	if !errors.Is(err, ErrSetterEOF) {
		return false, err
	}

	// Nested setter fully exhausted: disable it and advance
	// pointer for using the following field in this setter.
	s.nested = nil
	s.curFld++
	return true, s.checkEOF()
}

// doCurrent performs the operation against the current field.
func (s *primitiveSetter) doCurrent(op operation, value interface{}) (err error) {
	var fld interface{}
	if fld, err = s.getCurrentField(); err != nil {
		return
	}
	// Only opSet makes sense to execute. OpSkip always succeeds.
	if op == opSet {
		fldValue := reflect.ValueOf(fld)
		switch fldValue.Kind() {
		case reflect.Ptr:
			err = s.setPointerElem(fldValue.Elem(), value)
		default:
			err = fmt.Errorf("type %s is not supported", fldValue.Kind())
		}
	}
	return err
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
// field is already a Setter. Otherwise, nil is returned.
func (s *primitiveSetter) asSetter(fld interface{}) Setter {
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
