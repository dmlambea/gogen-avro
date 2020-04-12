package setters

import (
	"fmt"
	"reflect"
)

type sliceSetter struct {
	exhaustNotifierComponent
	entries   int
	curEntry  int
	slicePtr  reflect.Value
	sliceType reflect.Type
	item      reflect.Value
	inner     Setter
}

func newSliceSetter(sliceAddr interface{}, sliceType reflect.Type) Setter {
	s := &sliceSetter{
		slicePtr:  reflect.ValueOf(sliceAddr),
		sliceType: sliceType,
	}
	return s
}

// Init initializes the setter. The value argument is expected to be the item count
// to be consumed before getting exhausted.
func (s *sliceSetter) Init(arg interface{}) (err error) {
	var ok bool
	if s.entries, ok = arg.(int); !ok {
		return fmt.Errorf("wrong init argument type %t: expected int", arg)
	}

	if s.entries == 0 {
		return
	}

	newSlice := reflect.MakeSlice(s.sliceType, s.entries, s.entries)
	if !s.slicePtr.Elem().IsNil() {
		newSlice = reflect.AppendSlice(s.slicePtr.Elem(), newSlice)
	}
	s.slicePtr.Elem().Set(newSlice)
	return
}

// Execute should only be called for slice setters whenever value is actually the
// full contents of the fields.
func (s *sliceSetter) Execute(op OperationType, value interface{}) (err error) {
	if op != SetField {
		return
	}
	valueElem := reflect.ValueOf(value)
	if valueElem.Kind() != reflect.Slice && valueElem.Kind() != reflect.Array {
		return ErrTypeNotSupported
	}
	switch s.slicePtr.Elem().Kind() {
	case reflect.Slice:
		s.slicePtr.Elem().Set(valueElem)
	case reflect.Array:
		reflect.Copy(s.slicePtr.Elem(), valueElem)
	}
	s.entries = 0 // Intentionally exhaust this setter
	if s.hasExhaustCallback() {
		s.trigger(s)
	}
	return
}

// IsExhausted returns true if no more entries are expected to be consumed
func (s *sliceSetter) IsExhausted() bool {
	return s.entries <= 0
}

// GetInner creates the real memory for the map and the inner setter holding the
// key/value pairs for each consumption. This inner setter is asked for a notification
// every time it gets exhausted, so the key/val can be added to the map.
func (s *sliceSetter) GetInner() (inner Setter, err error) {
	s.item = reflect.New(s.sliceType.Elem())
	s.inner = NewSetterForFields([]interface{}{
		s.item.Elem().Addr().Interface(),
	})

	s.inner.setExhaustCallback(func(_ Setter) {
		s.callbackEvent()
	})
	return s.inner, err
}

func (s *sliceSetter) callbackEvent() {
	if s.inner.IsExhausted() {
		s.slicePtr.Elem().Index(s.curEntry).Set(s.item.Elem())
		s.curEntry++
		s.entries--
		if s.entries == 0 {
			s.trigger(s)
		}
		s.inner.(resettable).reset()
	}
}
