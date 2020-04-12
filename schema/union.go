package schema

import (
	"fmt"
	"strings"
)

const (
	unionNameFormat = "Union%s"
)

func NewUnionField(itemTypes []GenericType) *UnionType {
	t := &UnionType{}
	t.setItemTypes(itemTypes)
	unionName := t.generateUnionName()
	t.setFormatters(unionName, unionName)
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

	// Internally, optIndex is a positional index plus one of the null type within
	// an optional union in order to keep zero-value useful (0 = non-optional union).
	optIndex int
}

// generateUnionName must return a generated name for unions, which depends on its child types
func (t *UnionType) generateUnionName() string {
	var str strings.Builder
	for _, item := range t.itemTypes {
		str.WriteString(item.Name())
	}
	return fmt.Sprintf(unionNameFormat, str.String())
}

// IsOptional returns true if this union has a null-type option
func (t *UnionType) IsOptional() bool {
	return t.optIndex > 0
}

// IsSimple returns true if this union is optional and has only one another type
func (t *UnionType) IsSimple() bool {
	return t.IsOptional() && len(t.Children()) == 2
}

// SetOptionalIndex marks the positional index of the null type
func (t *UnionType) SetOptionalIndex(idx int) {
	t.optIndex = idx + 1
}

// OptionalIndex returns the index of the null type
func (t *UnionType) OptionalIndex() int {
	return t.optIndex - 1
}

// NonOptionalIndex has meaning on simple unions and returns the index of the non-null type.
func (t *UnionType) NonOptionalIndex() int {
	return 1 - (t.optIndex - 1)
}

func (t *UnionType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

func (t *UnionType) IsReadableBy(other GenericType, visited VisitMap) bool {
	u, otherIsUnion := other.(*UnionType)

	// Check the optional case for both unions
	if otherIsUnion && t.IsOptional() {
		return true
	}

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
