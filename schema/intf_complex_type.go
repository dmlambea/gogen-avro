package schema

// ComplexType is implemented by any AVRO type defined as "complex" in the specs:
//  fixem, enum, map, array, record and union
type ComplexType interface {
	GenericType
	isComplex() bool
}
