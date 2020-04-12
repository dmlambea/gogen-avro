package schema

type VisitMap map[string]bool

type GenericType interface {
	Name() string
	GoType() string
	SerializerMethod() string
	IsReadableBy(other GenericType, visited VisitMap) bool
}
