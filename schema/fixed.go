package schema

import "fmt"

func NewFixedField(qname QName, sizeBytes uint64) *FixedType {
	t := &FixedType{}
	t.setQName(qname)
	t.goType = fmt.Sprintf("[%d]byte", sizeBytes)
	t.sizeBytes = sizeBytes
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
	goType    string
	sizeBytes uint64
}

func (t *FixedType) SizeBytes() uint64 {
	return t.sizeBytes
}

func (t *FixedType) GoType() string {
	return t.goType
}

func (t *FixedType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

func (t *FixedType) IsReadableBy(other GenericType, visited VisitMap) bool {
	f, ok := other.(*FixedType)
	if ok {
		ok = (f.goType == t.goType) && (f.Name() == t.Name())
	}
	return ok
}

func (t *FixedType) isComplex() bool {
	return true
}
