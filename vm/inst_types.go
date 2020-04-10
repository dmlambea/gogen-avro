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
	// TODO: remove TypeNull as it will never get its way into the VM
	TypeNull
	TypeBool
	TypeInt
	TypeLong
	TypeFloat
	TypeDouble
	TypeString
	TypeBytes
	TypeFixed

	TypeAcc    // This type is special for moving accumulator only
	TypeRecord // This type is special for discarding record types only
	TypeBlock  // This type is special for discarding blocks types only
)

// TypeFromString is a utility function to get a primitive Type from its name
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
	case "fixed":
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
	case TypeFixed:
		return "fixed"
	case TypeAcc:
		return "acc"
	case TypeRecord:
		return "record"
	case TypeBlock:
		return "block"
	default:
		return fmt.Sprintf("<invalid type %d>", t)
	}
}
