package schema

// NewPrimitiveType makes and/or returns the cached definition of a primitive type.
func NewPrimitiveType(avroName string) *primitiveType {
	return primitiveTypesCache[avroName]
}

// Fixture used to create all primitive types and their compatibility list.
// All types are compatible with themselves.
type primitiveTypeFixture struct {
	avroName   string
	typeName   string
	goTypeName string
	compat     []string
}

var (
	primitiveTypesCache = createPrimitiveTypesMap([]primitiveTypeFixture{
		primitiveTypeFixture{avroName: "int", typeName: "Int", goTypeName: "int32", compat: []string{"int", "long", "float", "double"}},
		primitiveTypeFixture{avroName: "long", typeName: "Long", goTypeName: "int64", compat: []string{"long", "float", "double"}},
		primitiveTypeFixture{avroName: "float", typeName: "Float", goTypeName: "float32", compat: []string{"float", "double"}},
		primitiveTypeFixture{avroName: "double", typeName: "Double", goTypeName: "float64", compat: []string{"double"}},
		primitiveTypeFixture{avroName: "boolean", typeName: "Bool", goTypeName: "bool", compat: []string{"boolean"}},
		primitiveTypeFixture{avroName: "bytes", typeName: "Bytes", goTypeName: "[]byte", compat: []string{"bytes", "string"}},
		primitiveTypeFixture{avroName: "string", typeName: "String", goTypeName: "string", compat: []string{"string", "bytes"}},
		primitiveTypeFixture{avroName: "null", typeName: "Null", goTypeName: "", compat: []string{"null"}},
	})
)

// createPrimitiveTypesMap makes the cached definition of all primitive types.
func createPrimitiveTypesMap(fixtures []primitiveTypeFixture) map[string]*primitiveType {
	m := make(map[string]*primitiveType)
	for _, f := range fixtures {
		t := &primitiveType{name: f.typeName, goType: f.goTypeName, compat: f.compat}
		m[f.avroName] = t
	}
	return m
}

// Common attributes for all types
type primitiveType struct {
	optionalComponent
	name   string
	goType string
	compat []string
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

func (p *primitiveType) IsReadableBy(other GenericType, visited map[string]bool) bool {
	// If both fields are optional, they are compatible
	if p.IsOptional() && other.IsOptional() {
		return true
	}

	for _, compTypeName := range p.compat {
		if ct := NewPrimitiveType(compTypeName); ct != nil {
			if other.GoType() == ct.GoType() {
				return true
			}
		}
	}

	if u, ok := other.(*UnionType); ok {
		for _, child := range u.Children() {
			// Union children are fields, so their types is what is needed to match
			childType := child.(*FieldType).Type()
			if p.IsReadableBy(childType, visited) {
				return true
			}
		}
	}

	return false
}
