package schema

import (
	"fmt"
	"strings"
)

func NewEnumField(qname QName, symbols []string) *EnumType {
	t := &EnumType{symbols: make([]string, len(symbols))}
	for i := range symbols {
		t.symbols[i] = strings.Title(symbols[i])
	}
	t.setQName(qname)
	return t
}

var (
	// Ensure interface implementations
	_ ComplexType    = &EnumType{}
	_ DocumentedType = &EnumType{}
	_ NamespacedType = &EnumType{}
)

type EnumType struct {
	namespaceComponent
	optionalComponent
	documentComponent
	symbols []string
}

func (t *EnumType) GoType() string {
	return "int32"
}

func (t *EnumType) SerializerMethod() string {
	return fmt.Sprintf("write%s", t.Name())
}

func (t *EnumType) Symbols() []string {
	return t.symbols
}

func (t *EnumType) IsReadableBy(other GenericType, visited VisitMap) bool {
	f, ok := other.(*EnumType)
	if ok {
		ok = (f.Name() == t.Name())
	}
	return ok
}

func (t *EnumType) isComplex() bool {
	return true
}
