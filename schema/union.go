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
