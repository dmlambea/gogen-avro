package setters

import (
	"fmt"
	"reflect"
)

func newFieldListSetter(fields []interface{}) *fieldListSetter {
	origFields := make([]interface{}, len(fields))
	copy(origFields, fields)
	return &fieldListSetter{
		fldCount:   len(fields),
		origFields: origFields,
		fields:     fields,
	}
}

type fieldListSetter struct {
	exhaustNotifierComponent
	curFld     int
	fldCount   int
	origFields []interface{}
	fields     []interface{}
}

func (s *fieldListSetter) reset() (err error) {
	s.curFld = 0
	for i := range s.fields {
		r, ok := s.fields[i].(resettable)
		switch ok {
		case true:
			if err = r.reset(); err != nil {
				return
			}
		case false:
			s.fields[i] = s.origFields[i]
		}
	}
	return nil
}

// Primitive setters allow for no initialization.
// TODO use this to infor about the order of fields.
func (s *fieldListSetter) Init(arg interface{}) error {
	positions, ok := arg.([]int)
	if !ok {
		return fmt.Errorf("struct setter initialization expects []int, got %T", positions)
	}
	return s.sort(positions)
}

func (s *fieldListSetter) IsExhausted() bool {
	return s.curFld >= s.fldCount
}

// GetInner returns the current field, which must be a setter, or an error otherwise.
// This method changes the rules of the game for that field by:
//  a) initializing the target field type, if required
//  b) installing a notification callback in order to know when to advance the current
//     field pointer.
func (s *fieldListSetter) GetInner() (inner Setter, err error) {
	var field interface{}
	if field, err = s.getCurrentField(); err != nil {
		return
	}

	// Not a setter? Try create one
	var ok bool
	if inner, ok = field.(Setter); !ok {
		inner, err = createPointerToStructFieldSetter(field)
		if err != nil {
			return
		}
		// Settle down setter
		s.fields[s.curFld] = inner
	}

	// Must install a notification cb to be informed of field consumption
	inner.setExhaustCallback(func(_ Setter) {
		s.goNext()
	})
	return
}

// Execute performs the given operation and advances its current field pointer,
// if apply. If the current field happens to be a nested field, the operation is applied
// recursively. In this case, the current pointer cannot be advanced until the inner
// setter gets exhausted. Executing past the last field returns ErrSetterEOF.
func (s *fieldListSetter) Execute(op OperationType, value interface{}) (err error) {
	var fld interface{}
	if fld, err = s.getCurrentField(); err != nil {
		return
	}

	// If there is a nested Setter, use it for this field. Nested triggers goNext when exhausted.
	if inner, ok := fld.(Setter); ok {
		return s.executeNested(inner, op, value, fld)
	}

	// Only SetField makes sense to execute. SkipField always succeeds.
	err = nil
	if op == SetField {
		fldValue := reflect.ValueOf(fld)
		switch fldValue.Kind() {
		case reflect.Ptr:
			err = s.setPointerElem(fldValue.Elem(), value)
		default:
			err = ErrTypeNotSupported
		}
	}
	if err == nil {
		s.goNext()
	}
	return
}

// executeNested tries to match current field as Setter/pointer-to-Setter, then executes
// the operation in it. If not a Setter, ErrNotSetter is returned.
func (s *fieldListSetter) executeNested(inner Setter, op OperationType, value, field interface{}) (err error) {
	if !inner.hasExhaustCallback() {
		inner.setExhaustCallback(func(_ Setter) {
			s.goNext()
		})
	}
	return inner.Execute(op, value)
}

// sort replaces the ordering of the fields within this setter. the length of positions
// cannot exceed the length of the fields array. The position indexes must be between 0
// and len(fields)-1. All fields not referred to in the positions array are put in order
// of appearance at the end of the list.
func (s *fieldListSetter) sort(positions []int) (err error) {
	sortedFields := make([]interface{}, s.fldCount)
	for i := range positions {
		sortedFields[i] = s.fields[positions[i]]
		s.fields[positions[i]] = nil
	}
	posCount := len(positions)
	if posCount != s.fldCount {
		for i := 0; i < s.fldCount; i++ {
			if s.fields[i] != nil {
				sortedFields[posCount] = s.fields[i]
				posCount++
				if posCount == s.fldCount {
					break
				}
			}
		}
	}
	s.fields = sortedFields
	return
}

// goNext advances internal current pointer. No error checking is performed.
func (s *fieldListSetter) goNext() {
	s.curFld++
	if s.IsExhausted() && s.hasExhaustCallback() {
		s.trigger(s)
	}
}

// getCurrentField gets the current field to be set.
// Overflow condition is checked and returned as error.
func (s *fieldListSetter) getCurrentField() (fld interface{}, err error) {
	switch {
	case s.IsExhausted():
		err = ErrExhausted
	default:
		fld = s.fields[s.curFld]
	}
	return
}

// setPointerElem sets the value into the current field, which must be a pointer.
func (s *fieldListSetter) setPointerElem(elem reflect.Value, value interface{}) error {
	// If the target elem is an indirection, its pointer must be created first and it could
	// probably become a Setter.
	innerSetter, err := createPointerToStructReflectValueSetter(elem)
	switch err {
	case nil:
		// Make setter final
		s.fields[s.curFld] = innerSetter
		return s.executeNested(innerSetter, SetField, value, innerSetter)
	case ErrNotSetter:
		// Point to the target item type, then wait for the Set below
		elem = elem.Elem()
	case ErrNonPointer:
		// Nothing, just wait for the Set below
	default:
		return err
	}
	elem.Set(reflect.ValueOf(value))
	return nil
}

func createPointerToStructFieldSetter(field interface{}) (inner Setter, err error) {
	elem := reflect.ValueOf(field).Elem()
	return createPointerToStructReflectValueSetter(elem)
}

func createPointerToStructReflectValueSetter(elem reflect.Value) (inner Setter, err error) {
	switch elem.Type().Kind() {
	case reflect.Ptr:
		// Let it go further
	case reflect.Map, reflect.Struct:
		return NewSetterFor(elem.Addr().Interface())
	default:
		return nil, ErrNonPointer
	}

	// This creates the target field type
	indirectElem := reflect.New(elem.Type().Elem())
	elem.Set(indirectElem)

	// Try to detect a Setter from it
	inner, err = NewSetterFor(indirectElem.Interface())
	if err != nil {
		return nil, ErrNotSetter
	}
	return
}
