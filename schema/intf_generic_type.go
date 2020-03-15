package schema

type GenericType interface {
	OptionalType
	Name() string
	GoType() string
	SerializerMethod() string
	IsReadableBy(other GenericType, visited map[string]bool) bool
}
