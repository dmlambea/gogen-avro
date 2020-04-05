package setters

import (
	"errors"
	"fmt"
	"reflect"
)

type sliceSetter struct {
	exhaustNotifierComponent
	entries   int
	curEntry  int
	sliceElem reflect.Value
	sliceType reflect.Type
	item      reflect.Value
	inner     Setter
}

func newSliceSetter(sliceAddr interface{}, sliceType reflect.Type) Setter {
	s := &sliceSetter{
		sliceElem: reflect.ValueOf(sliceAddr),
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
	if !s.sliceElem.Elem().IsNil() {
		newSlice = reflect.AppendSlice(s.sliceElem.Elem(), newSlice)
	}
	s.sliceElem.Elem().Set(newSlice)
	return
}

// Execute should not be called for map setters. The inner setter should be
// used instead.
func (s *sliceSetter) Execute(op OperationType, value interface{}) (err error) {
	return errors.New("shouldn't be called directly")
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
		s.sliceElem.Elem().Index(s.curEntry).Set(s.item.Elem())
		s.curEntry++
		s.entries--
		if s.entries == 0 {
			s.trigger(s)
		}
		s.inner.(resettable).reset()
	}
}