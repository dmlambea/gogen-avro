package setters

import (
	"errors"
	"fmt"
	"reflect"
)

type mapKeyVal struct {
	K reflect.Value
	V reflect.Value
}

type mapSetter struct {
	exhaustNotifierComponent
	entries      int
	mapAddr      interface{}
	mapType      reflect.Type
	mapElem      reflect.Value
	keyVal       *mapKeyVal
	keyValSetter Setter
}

var mapIdx int

func newMapSetter(mapAddr interface{}, mapType reflect.Type) Setter {
	mapIdx++
	s := &mapSetter{mapAddr: mapAddr, mapType: mapType}
	return s
}

// Init initializes the setter. The value argument is expected to be the item count
// to be consumed before getting exhausted.
func (s *mapSetter) Init(arg interface{}) (err error) {
	var ok bool
	if s.entries, ok = arg.(int); !ok {
		err = fmt.Errorf("wrong init argument type %t: expected int", arg)
	}
	return
}

// Execute should not be called for map setters. The inner setter should be
// used instead.
// TODO optimize the usage of maps by allowing direct calls to Execute
func (s *mapSetter) Execute(op OperationType, value interface{}) (err error) {
	return errors.New("shouldn't be called directly")
}

// IsExhausted returns true if no more entries are expected to be consumed
func (s *mapSetter) IsExhausted() bool {
	return s.entries <= 0
}

// GetInner creates the real memory for the map and the inner setter holding the
// key/value pairs for each consumption. This inner setter is asked for a notification
// every time it gets exhausted, so the key/val can be added to the map.
func (s *mapSetter) GetInner() (inner Setter, err error) {
	s.mapElem = reflect.MakeMap(s.mapType)
	reflect.ValueOf(s.mapAddr).Elem().Set(s.mapElem)

	s.keyVal = &mapKeyVal{
		K: reflect.New(s.mapType.Key()),
		V: reflect.New(s.mapType.Elem()),
	}
	s.keyValSetter = NewSetterForFields([]interface{}{
		s.keyVal.K.Elem().Addr().Interface(),
		s.keyVal.V.Elem().Addr().Interface(),
	})

	s.keyValSetter.setExhaustCallback(func(_ Setter) {
		receiveNotification(s)
	})
	return s.keyValSetter, err
}

func receiveNotification(m *mapSetter) {
	if m.keyValSetter.IsExhausted() {
		m.mapElem.SetMapIndex(m.keyVal.K.Elem(), m.keyVal.V.Elem())
		m.entries--
		if m.entries == 0 {
			m.trigger(m)
		}
		m.keyValSetter.(resettable).reset()
	}
}
