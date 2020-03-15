package schema

import "fmt"

/*
  A named Reference to another type (fixed, enum or record). Just a wrapper with a qname around a qnamed type.
  If a reference is created without a defined type, all parsing-time values are returned as invalid values.
  Once the reference gets assigned a type, it triggers its registered resolvers' Resolve(ref) method.
  This is commonly transparent for almost all types, but some other types depend on their children fields for
  defining their internal data, like Name and GoType.
  Examples of this are:
	- Arrays of base types defined as references: the resolver renames the array to its final name.
	- Union types with named-types' members: once all its children refs are triggered, the union gets renamed.
*/
type Reference struct {
	qname     QName
	optIndex  int
	refType   GenericType
	resolvers []ReferenceResolver
}

// ReferenceResolver is an interface with a function that triggers once this
// reference gets informed about the type it refers to, so no afterwards resolving
// phase is required.
type ReferenceResolver interface {
	Resolve(ref Reference)
}

var (
	// Ensure interface implementations
	_ GenericType = &Reference{}
)

func NewReference(qname QName, t GenericType) *Reference {
	return &Reference{
		qname:   qname,
		refType: t,
	}
}

func (r Reference) Name() string {
	return r.qname.String()
}

func (r Reference) GoType() string {
	if r.refType == nil {
		return "untyped"
	}
	return r.refType.GoType()
}

func (r Reference) QName() QName {
	return r.qname
}

func (r *Reference) SerializerMethod() string {
	panic(fmt.Sprintf("This reference %T should have been resolved before", r))
}

func (r *Reference) IsOptional() bool {
	panic("Can references hold an optional index??")
}

func (r *Reference) IsUnion() bool {
	panic("Can references hold an optional index??")
}

func (r *Reference) SetOptionalIndex(idx int) {
	panic("Can references hold an optional index??")
}

func (r *Reference) OptionalIndex() int {
	panic("Can references hold an optional index??")
}

func (r *Reference) NonOptionalIndex() int {
	panic("Can references hold an optional index??")
}

func (r Reference) IsUntyped() bool {
	return r.refType == nil
}

func (r *Reference) Type() GenericType {
	return r.refType
}

func (r *Reference) SetType(t GenericType) {
	if r.refType != nil {
		panic("Cannot reassign reference type")
	}
	r.refType = t

	for _, resolver := range r.resolvers {
		resolver.Resolve(*r)
	}
}

func (r *Reference) AddResolver(resolver ReferenceResolver) {
	if resolver == nil {
		panic("Cannot add a nil resolver")
	}
	r.resolvers = append(r.resolvers, resolver)
}
