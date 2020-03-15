package schema

import "fmt"

const (
	mapNameAndTypeFormat = "Map%s"
)

func NewMapField(itemType GenericType) *MapType {
	t := &MapType{}
	t.setFormatters(mapNameAndTypeFormat, mapNameAndTypeFormat)
	t.setItemType(itemType)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType       = &MapType{}
	_ CompositeType     = &MapType{}
	_ ReferenceResolver = &MapType{}
)

type MapType struct {
	singleChildComponent
}

func (t *MapType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}
