package setters

import (
	"errors"
	"reflect"
)

// OperationType is the type for specifying the operation to be executed by a setter
type OperationType byte

const (
	// SetField means the requested operation is to set the current field
	SetField OperationType = iota

	// SkipField means the requested operation is to skip the current field
	SkipField
)

// Setter is the interface that must be implemented by all
// top-down data setters
type Setter interface {
	Init(arg interface{}) error
	Execute(op OperationType, value interface{}) error
	IsExhausted() bool
	GetInner() (Setter, error)

	setExhaustCallback(eventFunc)
	hasExhaustCallback() bool
}

// resettable is implemented by setters that can be reused
type resettable interface {
	reset() error
}

var (
	// ErrExhausted is a sentinel error indicating that all the
	// fields have been set, so this setter is exhausted.
	ErrExhausted = errors.New("setter exhausted")

	// ErrNotSetter is a sentinel error indicating that field is not a Setter.
	// This error is raised when calling GetInner over a non-setter current field.
	ErrNotSetter = errors.New("not a Setter")

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

	// ErrTypeNotSupported is a sentinel error for unsupported types
	ErrTypeNotSupported = errors.New("unsupported type")
)

// NewSetterForFields creates a setter for the given field addresses/setters.
func NewSetterForFields(fields []interface{}) Setter {
	processedFields := make([]interface{}, len(fields))
	for i := range fields {
		setter, err := NewSetterFor(fields[i])
		switch err {
		case nil:
			processedFields[i] = setter
		default:
			processedFields[i] = fields[i]
		}
	}
	return newFieldListSetter(processedFields)
}

// NewSetterFor creates a setter for the given type pointed by object.
func NewSetterFor(object interface{}) (Setter, error) {
	if s, ok := isAlreadySetter(object); ok {
		return s, nil
	}

	ptr, err := assertPointer(object)
	if err != nil {
		return nil, err
	}

	elem := ptr.Elem()
	switch elem.Kind() {
	case reflect.Struct:
		return newStructSetter(elem.Addr().Interface())
	case reflect.Map:
		return newMapSetter(object, elem.Type()), nil
	}
	return nil, ErrTypeNotSupported
}

// isAlreadySetter returns the (already-Setter, true) value of object, or (invalid, false) otherwise.
func isAlreadySetter(object interface{}) (s Setter, ok bool) {
	if object != nil {
		// Check if object is already a Setter
		if s, ok = object.(Setter); ok {
			return
		}
	}
	return nil, false
}

// newStructSetter creates a setter for all exported fields in struct elem.
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
		/*
			switch fld.Kind() {
			case reflect.Struct:
				if iface, err = NewSetterFor(iface); err != nil {
					return
				}
			case reflect.Map:
				//iface = newMapSetter(iface, fld.Type())
			default:
				// As is
			}
		*/
		setterFields = append(setterFields, iface)
	}
	return newFieldListSetter(setterFields), nil
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