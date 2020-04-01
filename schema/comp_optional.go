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

func (comp *optionalComponent) IsOptional() bool {
	return comp.optIndex > 0
}

func (comp *optionalComponent) IsUnion() bool {
	return false
}

func (comp *optionalComponent) SetOptionalIndex(idx int) {
	comp.optIndex = idx + 1
}

func (comp *optionalComponent) OptionalIndex() int {
	return comp.optIndex - 1
}

func (comp *optionalComponent) NonOptionalIndex() int {
	return 1 - (comp.optIndex - 1)
}
