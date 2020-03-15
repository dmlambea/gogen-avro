package schema

var (
	primitiveTypes = make(map[string]*primitiveType)
)

// NewPrimitiveType makes and/or returns the cached definition of a primitive type.
func NewPrimitiveType(name, goType string) *primitiveType {
	t, ok := primitiveTypes[goType]
	if !ok {
		t = &primitiveType{name: name, goType: goType}
		primitiveTypes[goType] = t
	}
	return t
}

// Common attributes for all types
type primitiveType struct {
	optionalComponent
	name   string
	goType string
}

var (
	// Ensure interface implementation
	_ GenericType  = &primitiveType{}
	_ OptionalType = &primitiveType{}
)

func (p *primitiveType) Name() string {
	return p.name
}

func (p *primitiveType) GoType() string {
	return p.goType
}

func (p *primitiveType) setGoType(goType string) {
	p.goType = goType
}

func (p *primitiveType) SerializerMethod() string {
	return "vm.WritePrimitive"
}
