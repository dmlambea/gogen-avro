package schema

import "fmt"

func NewRecordField(qname QName, itemTypes []GenericType) *RecordType {
	t := &RecordType{}
	t.setFormatters("RecNameLeches%s", "RecTypeLeches%s")
	t.setQName(qname)
	t.setItemTypes(itemTypes)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType    = &RecordType{}
	_ CompositeType  = &RecordType{}
	_ DocumentedType = &RecordType{}
	_ NamespacedType = &RecordType{}
)

type RecordType struct {
	namespaceComponent
	documentComponent
	multiChildComponent
}

// Disambiguate Name method from ComplexType and NamespacedType
func (t *RecordType) Name() string {
	return DefaultNamer.ToPublicName(t.QName().String())
}

func (t *RecordType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

/*
func (t *RecordType) Name() string {
	return t.GoType()
}

func (t *RecordType) GoType() string {
	return DefaultNamer.ToPublicName(t.QName().String())
}

func (t *RecordType) Children() []GenericType {
	return t.fieldsAsChildren
}
*/
