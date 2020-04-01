package schema

import (
	"fmt"
	"strings"
)

var (
	// Ensure interface implementations
	_ ComplexType       = &multiChildComponent{}
	_ CompositeType     = &multiChildComponent{}
	_ ReferenceResolver = &multiChildComponent{}
)

type multiChildComponent struct {
	optionalComponent
	nameFmt   string
	goTypeFmt string
	itemTypes []GenericType
}

// *** Generic type implementation

func (t *multiChildComponent) Name() string {
	t.avoidDerreferencingUnresolvedRefs()
	return t.formatItemID(t.nameFmt, func(item GenericType) string {
		return item.Name()
	})
}

func (t *multiChildComponent) GoType() string {
	t.avoidDerreferencingUnresolvedRefs()
	return t.formatItemID(t.goTypeFmt, func(item GenericType) string {
		return item.GoType()
	})
}

func (t *multiChildComponent) SerializerMethod() string {
	panic("Complex, multi-child types must implement their own SerializerMethod")
}

func (t *multiChildComponent) IsReadableBy(other GenericType, visited VisitMap) bool {
	if t.GoType() == other.GoType() {
		return true
	}

	if u, ok := other.(*UnionType); ok {
		for _, child := range u.Children() {
			// Union children are fields, so their types is what is needed to match
			childType := child.(*FieldType).Type()
			if t.IsReadableBy(childType, visited) {
				return true
			}
		}
	}
	return false
}

// *** Complex type implementation

func (t *multiChildComponent) isComplex() bool {
	return true
}

// *** Composite type implementation

func (t *multiChildComponent) Children() []GenericType {
	return t.itemTypes
}

// *** Reference resolver implementation

func (t *multiChildComponent) Resolve(ref Reference) {
	for i, item := range t.itemTypes {
		if itemRef, ok := item.(*Reference); ok {
			if itemRef.QName().String() == ref.QName().String() {
				t.itemTypes[i] = ref.Type()
				return
			}
		}
	}
	panic("Reference not found in children")
}

// *** Internal tooling

// setFormatters fix the formatting strings used to make the name and go type name when requested
func (t *multiChildComponent) setFormatters(nameFmt, goTypeFmt string) {
	t.nameFmt = nameFmt
	t.goTypeFmt = goTypeFmt
}

func (t *multiChildComponent) setItemTypes(itemTypes []GenericType) {
	t.itemTypes = itemTypes
	for _, item := range t.itemTypes {
		if ref, ok := item.(*Reference); ok {
			ref.AddResolver(t)
		}
	}
}

func (t *multiChildComponent) formatItemID(fmtStr string, idGetter func(item GenericType) string) string {
	var str strings.Builder

	for _, item := range t.itemTypes {
		str.WriteString(idGetter(item))
	}
	if fmtStr == "" {
		return str.String()
	}
	return fmt.Sprintf(fmtStr, str.String())
}

// Checks that there are no unresolved references among the children
func (t *multiChildComponent) avoidDerreferencingUnresolvedRefs() {
	for _, item := range t.itemTypes {
		if ref, ok := item.(*Reference); ok {
			panic(fmt.Sprintf("Unresolved reference '%s'", ref.Name()))
		}
	}
}
