package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompileSimple(t *testing.T) {
	schemas := []string{
		`{ "name": "Int", "type": "int" }`,
		`{ "name": "Long", "type": "long" }`,
		`{ "name": "Float", "type": "float" }`,
		`{ "name": "Double", "type": "double" }`,
		`{ "name": "String", "type": "string" }`,
	}

	for _, s := range schemas {
		_, err := CompileSchemaBytes([]byte(s), []byte(s))
		require.Nil(t, err)
	}
}
