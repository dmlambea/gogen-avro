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

func (comp *singleChildComponent) Name() string {
	return fmt.Sprintf(comp.nameFmt, comp.Type().Name())
}

func (comp *singleChildComponent) GoType() string {
	return fmt.Sprintf(comp.goTypeFmt, comp.Type().GoType())
}

func (comp *singleChildComponent) SerializerMethod() string {
	return fmt.Sprintf("write%s", comp.Name())
}

func (comp *singleChildComponent) Type() GenericType {
	return comp.Children()[0]
}

func (comp *singleChildComponent) setItemType(itemType GenericType) {
	comp.multiChildComponent.setItemTypes([]GenericType{itemType})
}

func (comp *singleChildComponent) setItemTypes(itemTypes []GenericType) {
	panic(fmt.Sprintf("%v is a single-child type and it has a method for setting just one child type", comp))
}
