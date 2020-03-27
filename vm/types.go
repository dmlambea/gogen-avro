package vm

import (
	"fmt"
	"strings"
)

// Type defines the data type for a mov or skip opcode
type Type byte

// Constant values for Type
const (
	TypeError Type = iota
	TypeNull
	TypeBool
	TypeInt
	TypeLong
	TypeFloat
	TypeDouble
	TypeString
	TypeBytes
)

// TypeFromString is a utility function to get a Type from its name
func TypeFromString(name string) Type {
	switch strings.ToLower(name) {
	case "null":
		return TypeNull
	case "bool":
		return TypeBool
	case "int":
		return TypeInt
	case "long":
		return TypeLong
	case "float":
		return TypeFloat
	case "double":
		return TypeDouble
	case "string":
		return TypeString
	case "bytes":
		return TypeBytes
	default:
		return TypeError
	}
}

func (t Type) String() string {
	switch t {
	case TypeError:
		return "<error>"
	case TypeNull:
		return "null"
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeLong:
		return "long"
	case TypeFloat:
		return "float"
	case TypeDouble:
		return "double"
	case TypeString:
		return "string"
	case TypeBytes:
		return "bytes"
	default:
		return fmt.Sprintf("<invalid type %d>", t)
	}
}
