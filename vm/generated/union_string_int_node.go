package generated

import (
	"reflect"

	"github.com/actgardner/gogen-avro/vm/setters"
)

const (
	UnionStringIntNodeTypeString int64 = 1
	UnionStringIntNodeTypeInt    int64 = 2
	UnionStringIntNodeTypeNode   int64 = 3
)

type UnionStringIntNode struct {
	setters.BaseUnion
}

// Support method for deserialization
func (u UnionStringIntNode) UnionTypes() []reflect.Type {
	return unionStringIntNodeTypes
}

var (
	unionStringIntNodeTypes = []reflect.Type{
		reflect.TypeOf(nil), // Blank type for null
		reflect.TypeOf((*string)(nil)),
		reflect.TypeOf((*int32)(nil)),
		reflect.TypeOf((*Node)(nil)),
	}
)
