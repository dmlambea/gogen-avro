package schema

type OptionalType interface {
	IsOptional() bool
	IsUnion() bool // Reserved for unions, which might be optional and have several children, so no single non-optional index will exist
	SetOptionalIndex(idx int)
	OptionalIndex() int
	NonOptionalIndex() int
}

// Common attributes for all optional types
type optionalComponent struct {
	optIndex int // Positional index plus one of the null type within an optional union; zero-value = non-optional field
}

var (
	// Ensure interface implementation
	_ OptionalType = &optionalComponent{}
)

func (p *optionalComponent) IsOptional() bool {
	return p.optIndex > 0
}

func (p *optionalComponent) IsUnion() bool {
	return false
}

func (p *optionalComponent) SetOptionalIndex(idx int) {
	p.optIndex = idx + 1
}

func (p *optionalComponent) OptionalIndex() int {
	return p.optIndex - 1
}

func (p *optionalComponent) NonOptionalIndex() int {
	return 1 - (p.optIndex - 1)
}
