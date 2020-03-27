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

// Set puts the value into the current field this Setter is tracking.
// A successful set advances the current field. Setting past the last
// field returns ErrSetterEOF.
func (s *mapSetter) Set(value interface{}) (err error) {
	// Uninitialized map
	if s.keyValSetter == nil {
		return ErrUninitializedSetter
	}

	// Exhausted map
	if s.entries == 0 {
		return ErrSetterEOF
	}

	err = s.keyValSetter.Set(value)
	if err == ErrSetterEOF {
		// EOF is returned from keyVal setter, once exhausted
		s.appendMap()
		s.entries--
		if s.entries > 0 {
			// Reuse the keyVal setter for the rest of the entries
			s.keyValSetter.Reset()
			err = nil
		} else {
			// Remove the keyVal setter, since this map would need
			// initialization again, and populate EOF condition
			s.keyValSetter = nil
		}
	}
	// This point returns normal EOF condition for the map, or an error
	// from any of the nested setters.
	return
}

func (s *mapSetter) Reset() error {
	return ErrNotResettable
}

// appendMap puts the ky-val pair into the target map
func (s *mapSetter) appendMap() {
	k := s.keyVal.K.Elem()
	v := s.keyVal.V.Elem()
	s.mapElem.SetMapIndex(k, v)
}
