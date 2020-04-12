package schema

const (
	mapNameFormat = "Map%s"
	mapTypeFormat = "map[string]%s"
)

func NewMapField(itemType GenericType) *MapType {
	t := &MapType{}
	t.setFormatters(mapNameFormat, mapTypeFormat)
	t.setItemType(itemType)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType       = &MapType{}
	_ CompositeType     = &MapType{}
	_ SingleChildType   = &MapType{}
	_ ReferenceResolver = &MapType{}
)

type MapType struct {
	singleChildComponent
}

func (t *MapType) IsReadableBy(other GenericType, visited VisitMap) bool {
	if m, ok := other.(*MapType); ok {
		return t.Type().IsReadableBy(m.Type(), visited)
	}

	return t.singleChildComponent.IsReadableBy(other, visited)
}
