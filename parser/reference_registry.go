package parser

import (
	"fmt"

	"github.com/actgardner/gogen-avro/schema"
)

// referenceRegistry is a special registry for all qnamed types, which could be possibly referenced
// via aliases or type names elsewhere. The registry not only avoids duplicating references in memory,
// but also triggers the resolution of a reference, if it was created before defining their real type.
// By triggering the resolution, all types whose properties depend on their children or base types can
// be finally set.
type referenceRegistry struct {
	refs map[schema.QName]*schema.Reference
}

func NewReferenceRegistry() referenceRegistry {
	return referenceRegistry{
		refs: make(map[schema.QName]*schema.Reference),
	}
}

// CreateReference returns a Reference for a given qnamed type. If the reference already exists, its value is returned
// instead of creating a duplicated one. If the type being registered is unknown at registration time, the reference
// gets registered untyped. Once the real, final type is registered, the reference is updated using its SetType method.
// An untyped reference being setted this way triggers its resolution, allowing all its "owner" types to refresh their
// internal data, if needed.
func (r referenceRegistry) CreateReference(qname schema.QName, t schema.GenericType) (*schema.Reference, error) {
	ref := r.getOrCreateUntypedReference(qname)
	if t == nil {
		return ref, nil
	}

	if !ref.IsUntyped() {
		return nil, fmt.Errorf("Conflicting definitions for %v: type defined twice", qname)
	}

	// Trigger type update for reference
	ref.SetType(t)

	// Trigger type update for all of its aliases
	if nt, ok := t.(schema.NamespacedType); ok {
		for _, alias := range nt.Aliases() {
			aliasedRef := r.getOrCreateUntypedReference(alias)
			if !aliasedRef.IsUntyped() {
				return nil, fmt.Errorf("Alias %s from %s is conflicting with definitions for %s", alias, aliasedRef.Name(), ref.Name())
			}
			aliasedRef.SetType(t)
		}
	}
	return ref, nil
}

func (r referenceRegistry) getOrCreateUntypedReference(name schema.QName) *schema.Reference {
	ref := r.refs[name]
	if ref == nil {
		ref = schema.NewReference(name, nil)
		r.refs[name] = ref
	}
	return ref
}
