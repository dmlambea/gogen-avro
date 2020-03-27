package setter

import (
	"errors"
	"fmt"
	"reflect"
)

// Setter is the interface that must be implemented so that
// the VM can set a data stream over a structure.
type Setter interface {
	Init(arg interface{}) error
	Set(value interface{}) error
	Reset() error
}

// Factory is the interface for types that can create their
// own setters at runtime.
type Factory interface {
	Setter() Setter
}

var (
	// ErrSetterEOF is a sentinel error indicating that all the
	// fields have been set, so this setter is exhausted.
	ErrSetterEOF = errors.New("EOF setting field")

	// ErrUninitializedSetter is a sentinel error indicating that the
	// current setter required initialization before being used, like in map/arrays.
	ErrUninitializedSetter = errors.New("uninitialized setter")

	// ErrNoInitialization is a sentinel error indicating that the
	// leaf node setter requires no initialization.
	ErrNoInitialization = errors.New("setter requires no initialization")

	// ErrNotResettable is a sentinel error indicating this setter cannot be reused.
	ErrNotResettable = errors.New("setter cannot be reset")

	// ErrNil is returned when trying to create a setter to a nil address.
	ErrNil = errors.New("nil object")

	// ErrNonPointer arises when trying to create a setter from a value object.
	// Value objects disappear when the function returns, so a setter to them are useless.
	ErrNonPointer = errors.New("non-pointer object")

	// ErrNonStruct arises when trying to create a setter from an object type other than struct.
	ErrNonStruct = errors.New("non-struct object")
)

// NewSetterForFields creates a setter for the given field addresses/setters.
func NewSetterForFields(fields []interface{}) Setter {
	return &primitiveSetter{
		fields: fields,
	}
}

// NewSetterFor creates a setter for all exported fields in object.
func NewSetterFor(object interface{}) (Setter, error) {
	ptr, err := assertPointer(object)
	if err != nil {
		return nil, err
	}

	elem := ptr.Elem()
	switch elem.Kind() {
	/*
		case reflect.Ptr:
			if elem.CanAddr() && elem.Addr().CanInterface() {
				return NewSetterFor(elem.Addr().Interface())
			}
	*/
	case reflect.Struct:
		return newStructSetter(elem.Addr().Interface())
	case reflect.Map:
		panic("Does it occur at any time?")
		return newMapSetter(object, elem.Type()), nil
	}
	return nil, fmt.Errorf("object type %s not supported yet", elem.Kind())
}

// newStructSetter creates a setter for all exported fields in struct elem.
// If an interface is provided, it is first checked for Setter or Factory types first;
// otherwise, it is tried as interface{} generic with a real struct in it. If it all fails,
// an error is returned.
func newStructSetter(object interface{}) (s Setter, err error) {
	var ptr reflect.Value
	if ptr, err = assertPointer(object); err != nil {
		return
	}

	elem := ptr.Elem()
	total := elem.NumField()
	var setterFields []interface{}
	for i := 0; i < total; i++ {
		fld := elem.Field(i)
		if !fld.CanAddr() || !fld.Addr().CanInterface() {
			continue
		}
		iface := fld.Addr().Interface()
		switch fld.Kind() {
		case reflect.Struct:
			iface, err = NewSetterFor(iface)
			if err != nil {
				return
			}
		case reflect.Map:
			iface = newMapSetter(iface, fld.Type())
		default:
			// As is
		}
		setterFields = append(setterFields, iface)
	}
	return NewSetterForFields(setterFields), nil
}

func assertPointer(object interface{}) (ptr reflect.Value, err error) {
	if object == nil {
		err = ErrNil
	} else {
		ptr = reflect.ValueOf(object)
		if ptr.Kind() != reflect.Ptr {
			err = ErrNonPointer
		}
	}
	return
}
