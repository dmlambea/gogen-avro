package schema

import "strings"

func NewField(name string, itemType GenericType, index int) *FieldType {
	t := &FieldType{index: index}
	t.setQName(QName{Name: name})
	t.setItemType(itemType)
	return t
}

var (
	// Ensure interface implementations
	_ DocumentedType = &FieldType{}
	_ NamespacedType = &FieldType{}
)

type FieldType struct {
	documentComponent
	namespaceComponent
	singleChildComponent
	index int
}

func (t *FieldType) Index() int {
	return t.index
}

func (t *FieldType) HasDefault() bool {
	return false
}

func (t *FieldType) GoType() string {
	child := t.Type()
	var str strings.Builder
	if child.IsOptional() {
		str.WriteString("*")
	}

	if ct, ok := child.(ComplexType); ok {
		str.WriteString(DefaultNamer.ToPublicName(ct.Name()))
	} else {
		str.WriteString(child.GoType())
	}
	return str.String()
}

func (t *FieldType) SerializerMethod() string {
	return t.Type().SerializerMethod()
}

func (t *FieldType) IsOptional() bool {
	return t.Type().IsOptional()
}

func (t *FieldType) IsUnion() bool {
	return t.Type().IsUnion()
}

func (t *FieldType) OptionalIndex() int {
	return t.Type().OptionalIndex()
}

func (t *FieldType) NonOptionalIndex() int {
	return t.Type().NonOptionalIndex()
}

func (t *FieldType) IsReadableBy(other GenericType, visited VisitMap) bool {
	return t.Type().IsReadableBy(other, visited)
}

// TODO check if the names should always bo compared over their public versions
func (t *FieldType) alsoKnownAs(aka string) bool {
	if t.Name() == aka {
		return true
	}
	for _, alias := range t.Aliases() {
		if DefaultNamer.ToPublicName(alias.String()) == aka {
			return true
		}
	}
	return false
}
