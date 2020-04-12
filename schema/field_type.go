package schema

import (
	"fmt"
	"strings"
)

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

func (t *FieldType) Name() string {
	return DefaultNamer.ToPublicName(t.QName().String())
}

func (t *FieldType) GoType() string {
	child := t.Type()
	var str strings.Builder

	if t.IsOptional() {
		str.WriteString("*")
	}

	if t.IsSimple() {
		u := child.(*UnionType)
		str.WriteString(u.Children()[u.NonOptionalIndex()].GoType())
		return str.String()
	}

	if ct, ok := child.(ComplexType); ok {
		str.WriteString(DefaultNamer.ToPublicName(ct.Name()))
	} else {
		str.WriteString(child.GoType())
	}
	return str.String()
}

func (t *FieldType) SerializerMethod() string {
	if t.IsSimple() {
		u := t.Type().(*UnionType)
		return u.Children()[u.NonOptionalIndex()].SerializerMethod()
	}
	return t.Type().SerializerMethod()
}

func (t *FieldType) IsOptional() bool {
	u, ok := t.Type().(*UnionType)
	if !ok {
		return false
	}
	return u.IsOptional()
}

func (t *FieldType) IsSimple() bool {
	u, ok := t.Type().(*UnionType)
	if !ok {
		return false
	}
	return u.IsSimple()
}

func (t *FieldType) OptionalIndex() int {
	u, ok := t.Type().(*UnionType)
	if !ok {
		panic(fmt.Sprintf("field %s is not a union type", t.Name()))
	}
	return u.OptionalIndex()
}

func (t *FieldType) NonOptionalIndex() int {
	u, ok := t.Type().(*UnionType)
	if !ok {
		panic(fmt.Sprintf("field %s is not a union type", t.Name()))
	}
	return u.NonOptionalIndex()
}

func (t *FieldType) IsReadableBy(other GenericType, visited VisitMap) bool {
	if fld, ok := other.(*FieldType); ok {
		return t.Type().IsReadableBy(fld, visited)
	}
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
