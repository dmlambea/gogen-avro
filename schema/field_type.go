package schema

import "strings"

func NewField(name string, itemType GenericType) *FieldType {
	t := &FieldType{}
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
