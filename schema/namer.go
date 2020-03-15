package schema

import "strings"

// Namer is the interface defining a function for converting
// a name to a go-idiomatic public name.
type Namer interface {
	// ToPublicName returns a go-idiomatic public name.
	ToPublicName(name string) string
}

var (
	// Package-level namer for complex types
	DefaultNamer Namer = &defaultNamer{}
)

type defaultNamer struct{}

func (n *defaultNamer) ToPublicName(name string) string {
	return strings.Title(name)
}
