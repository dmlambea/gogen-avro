package schema

import "fmt"

func NewFixedField(qname QName, sizeBytes uint64) *FixedType {
	t := &FixedType{}
	t.setQName(qname)
	t.goType = fmt.Sprintf("[%d]byte", sizeBytes)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType    = &FixedType{}
	_ NamespacedType = &FixedType{}
)

type FixedType struct {
	namespaceComponent
	optionalComponent
	goType string
}

func (t *FixedType) GoType() string {
	return t.goType
}

func (t *FixedType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

func (t *FixedType) isComplex() bool {
	return true
}
