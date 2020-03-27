package setter

import (
	"fmt"
	"reflect"
)

type mapKeyVal struct {
	K reflect.Value
	V reflect.Value
}

type mapSetter struct {
	entries      int
	mapAddr      interface{}
	mapElem      reflect.Value
	mapType      reflect.Type
	keyVal       mapKeyVal
	keyValSetter Setter
}

func newMapSetter(mapAddr interface{}, mapType reflect.Type) Setter {
	s := &mapSetter{
		mapAddr: mapAddr,
		mapType: mapType,
	}
	return s
}

// Init initializes the setter. The value argument is expected to be the item count.
// This method is idempotent and can be called more than once, to allow more items to
// be consumed.
func (s *mapSetter) Init(arg interface{}) (err error) {
	// Remember that leaf nodes go first!
	if s.keyValSetter != nil {
		return s.keyValSetter.Init(arg)
	}

	count, ok := arg.(int)
	if !ok {
		return fmt.Errorf("wrong init argument type %t: expected int", arg)
	}
	s.entries = count
	// First-time initialization: create nested setter and the map itself
	if s.keyValSetter == nil {
		s.mapElem = reflect.MakeMap(s.mapType)
		reflect.ValueOf(s.mapAddr).Elem().Set(s.mapElem)
		s.keyVal.K = reflect.New(s.mapType.Key())
		s.keyVal.V = reflect.New(s.mapType.Elem())
		s.keyValSetter = NewSetterForFields([]interface{}{
			s.keyVal.K.Elem().Addr().Interface(),
			s.keyVal.V.Elem().Addr().Interface(),
		})
	}
	return
}

func (s *mapSetter) Set(value interface{}) error {
	return s.doOperation(opSet, value)
}

func (s *mapSetter) Skip() error {
	return s.doOperation(opSkip, nil)
}

func (s *mapSetter) Reset() error {
	return ErrNotResettable
}

func (s *mapSetter) doOperation(op operation, value interface{}) (err error) {
	// Uninitialized map
	if s.keyValSetter == nil {
		return ErrUninitializedSetter
	}

	// Exhausted map
	if s.entries == 0 {
		return ErrSetterEOF
	}

	switch op {
	case opSet:
		err = s.keyValSetter.Set(value)
	case opSkip:
		err = s.keyValSetter.Skip()
	}
	if err == ErrSetterEOF {
		// EOF is returned from keyVal setter, once exhausted
		s.appendMap()
		s.entries--
		if s.entries > 0 {
			// Reuse the keyVal setter for the rest of the entries
			s.keyValSetter.Reset()
			// Avoid populating EOF condition, since there are more entries to be operated.
			err = nil
		} else {
			// Remove the keyVal setter so that the next Init call affects the
			// mapSetter and not the nested keyValSetter
			s.keyValSetter = nil
		}
	}
	// This point returns normal EOF condition for the map, or an error
	// from any of the nested setters.
	return
}

// appendMap puts the ky-val pair into the target map
func (s *mapSetter) appendMap() {
	k := s.keyVal.K.Elem()
	v := s.keyVal.V.Elem()
	s.mapElem.SetMapIndex(k, v)
}
