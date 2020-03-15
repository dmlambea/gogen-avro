package schema

type GenericType interface {
	OptionalType
	Name() string
	GoType() string
	SerializerMethod() string
}
