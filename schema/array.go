package schema

import "fmt"

const (
	arrayNameFormat = "Array%s"
	arrayTypeFormat = "[]%s"
)

func NewArrayField(itemType GenericType) *ArrayType {
	t := &ArrayType{}
	t.setFormatters(arrayNameFormat, arrayTypeFormat)
	t.setItemType(itemType)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType       = &ArrayType{}
	_ CompositeType     = &ArrayType{}
	_ SingleChildType   = &ArrayType{}
	_ ReferenceResolver = &ArrayType{}
)

type ArrayType struct {
	singleChildComponent
}

func (t *ArrayType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}
