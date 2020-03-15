package schema

// CompositeType is implemented by any AVRO type able to contain children types,
// like arrays, maps, recods and unions.
type CompositeType interface {
	Children() []GenericType
}
