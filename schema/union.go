package schema

import "fmt"

const (
	unionNameFormat = "Union%s"
)

func NewUnionField(itemTypes []GenericType) *UnionType {
	t := &UnionType{}
	t.setFormatters(unionNameFormat, "UnionTypeLeches%s")
	t.setItemTypes(itemTypes)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType       = &UnionType{}
	_ CompositeType     = &UnionType{}
	_ ReferenceResolver = &UnionType{}
)

type UnionType struct {
	multiChildComponent
}

// Convenience function for telling the template this field is a union field, and therefore
// it might contain several non-optional children.
func (t *UnionType) IsUnion() bool {
	return true
}

func (t *UnionType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

func (t *UnionType) IsReadableBy(other GenericType, visited map[string]bool) bool {
	// Check the optional case, when the field is a null type
	if other.IsOptional() {
		return t.IsOptional()
	}

	u, otherIsUnion := other.(*UnionType)

	// Report if *any* writer type could be deserialized by the reader
	for _, child := range t.Children() {
		switch otherIsUnion {
		case true:
			for _, otherUnionChild := range u.Children() {
				// Union children are fields, so their types is what is needed to match
				otherUnionChildType := otherUnionChild.(*FieldType).Type()
				if child.IsReadableBy(otherUnionChildType, visited) {
					return true
				}
			}
		case false:
			if child.IsReadableBy(other, visited) {
				return true
			}
		}
	}
	return false
}
