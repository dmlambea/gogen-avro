package setters

import (
	"fmt"
	"reflect"
)

func newFieldListSetter(fields []interface{}) *fieldListSetter {
	return &fieldListSetter{
		sortableFieldsComponent: newSortableFieldsComponent(fields),
	}
}

type fieldListSetter struct {
	exhaustNotifierComponent
	sortableFieldsComponent
	currentField int
}

func (s *fieldListSetter) reset() (err error) {
	s.currentField = 0
	for i := range s.fields {
		if r, ok := s.fields[i].(resettable); ok {
			if err = r.reset(); err != nil {
				return
			}
		}
	}
	s.initSortOrder()
	return nil
}

// Field list setter can be initialized with the order its fields
// are to be read.
func (s *fieldListSetter) Init(arg interface{}) error {
	positions, ok := arg.([]int)
	if ok {
		s.sort(positions)
		return nil
	}
	return fmt.Errorf("struct setter initialization expects []int, got %T", positions)
}

func (s *fieldListSetter) IsExhausted() bool {
	return s.currentField >= s.fieldCount
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
		s.set(s.currentField, inner)
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
		return s.executeNested(op, value, inner)
	}

	// Only SetField makes sense to execute. SkipField always succeeds.
	var setterCreated bool
	err = nil
	if op == SetField {
		fldValue := reflect.ValueOf(fld)
		if fldValue.Kind() != reflect.Ptr {
			return ErrTypeNotSupported
		}
		setterCreated, err = s.setPointerElem(fldValue.Elem(), value)
	}
	if err == nil && !setterCreated {
		s.goNext()
	}
	return
}

// executeNested tries to match current field as Setter/pointer-to-Setter, then executes
// the operation in it. If not a Setter, ErrNotSetter is returned.
func (s *fieldListSetter) executeNested(op OperationType, value interface{}, inner Setter) (err error) {
	if !inner.hasExhaustCallback() {
		inner.setExhaustCallback(func(_ Setter) {
			s.goNext()
		})
	}
	return inner.Execute(op, value)
}

// goNext advances internal current pointer. No error checking is performed.
func (s *fieldListSetter) goNext() {
	s.currentField++
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
		fld = s.get(s.currentField)
	}
	return
}

// setPointerElem sets the value into the current field, which must be a pointer. Returns true
// if a new inner setter was created fo setting the field.
func (s *fieldListSetter) setPointerElem(elem reflect.Value, value interface{}) (bool, error) {
	// If the target elem is an indirection, its pointer must be created first and it could
	// probably become a Setter.
	innerSetter, err := createPointerToStructReflectValueSetter(elem)
	switch err {
	case nil:
		// Make setter final
		s.set(s.currentField, innerSetter)
		return true, s.executeNested(SetField, value, innerSetter)
	case ErrNotSetter:
		// Point to the target item type, then wait for the Set below
		elem = elem.Elem()
	case ErrNonPointer:
		// Nothing, just wait for the Set below
	default:
		return false, err
	}

	v := reflect.ValueOf(value)
	if !v.Type().AssignableTo(elem.Type()) {
		if !v.Type().ConvertibleTo(elem.Type()) {
			return false, fmt.Errorf("incompatible types: %s cannot be assigned to %s", v.Type(), elem.Type())
		}
		v = v.Convert(elem.Type())
	}
	elem.Set(v)
	return false, nil
}

func createPointerToStructFieldSetter(field interface{}) (inner Setter, err error) {
	elem := reflect.ValueOf(field).Elem()
	return createPointerToStructReflectValueSetter(elem)
}

func createPointerToStructReflectValueSetter(elem reflect.Value) (inner Setter, err error) {
	switch elem.Type().Kind() {
	case reflect.Ptr:
		// Let it go further
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
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
