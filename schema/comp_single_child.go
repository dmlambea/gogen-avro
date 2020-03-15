package schema

import (
	"fmt"
)

type SingleChildType interface {
	Type() GenericType
}

var (
	// Ensure interface implementations
	_ ComplexType       = &singleChildComponent{}
	_ CompositeType     = &singleChildComponent{}
	_ ReferenceResolver = &singleChildComponent{}
	_ SingleChildType   = &singleChildComponent{}
)

type singleChildComponent struct {
	multiChildComponent
}

func (t *singleChildComponent) SerializerMethod() string {
	panic("Complex, single-child types must implement their own SerializerMethod")
}

func (t *singleChildComponent) Type() GenericType {
	return t.Children()[0]
}

func (t *singleChildComponent) setItemType(itemType GenericType) {
	t.multiChildComponent.setItemTypes([]GenericType{itemType})
}

func (t *singleChildComponent) setItemTypes(itemTypes []GenericType) {
	panic(fmt.Sprintf("%v is a single-child type and it has a method for setting just one child type", t))
}
