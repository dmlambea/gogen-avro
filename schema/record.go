package schema

import "fmt"

func NewRecordField(qname QName, itemTypes []GenericType) *RecordType {
	t := &RecordType{}
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
	_ SchemaType     = &RecordType{}
)

type RecordType struct {
	namespaceComponent
	documentComponent
	multiChildComponent
	schemaComponent
}

// Disambiguate Name method from ComplexType and NamespacedType
func (t *RecordType) Name() string {
	return DefaultNamer.ToPublicName(t.QName().String())
}

func (t *RecordType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

func (t *RecordType) IsReadableBy(other GenericType, visited VisitMap) bool {
	// If there's a circular reference, don't evaluate every field on the second pass
	if _, ok := visited[t.Name()]; ok {
		return true
	}

	if otherRecord, ok := other.(*RecordType); ok {
		visited[t.Name()] = true

		for _, child := range otherRecord.Children() {
			readerField, ok := child.(*FieldType)
			if !ok {
				panic(fmt.Sprintf("Unexpected non-field type %T in %s", child, otherRecord.Name()))
			}
			writerField := t.FindFieldByNameOrAlias(readerField)

			// Two schemas are incompatible if the reader has a field with no default value that is not present in the writer schema
			if writerField == nil && !readerField.HasDefault() {
				return false
			}

			// The two schemas are incompatible if two fields with the same name have different schemas
			if writerField != nil && !writerField.Type().IsReadableBy(readerField.Type(), visited) {
				return false
			}
		}
		return true
	}

	return t.multiChildComponent.IsReadableBy(other, visited)
}

func (t *RecordType) FindFieldByNameOrAlias(sample *FieldType) *FieldType {
	for _, child := range t.Children() {
		field, ok := child.(*FieldType)
		if !ok {
			panic(fmt.Sprintf("Unexpected non-field type %T in %s", child, t.Name()))
		}

		if field.alsoKnownAs(sample.Name()) {
			return field
		}
		for _, alias := range sample.Aliases() {
			if field.alsoKnownAs(DefaultNamer.ToPublicName(alias.String())) {
				return field
			}
		}
	}
	return nil
}
