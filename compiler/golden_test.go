package compiler

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update .golden files")

func goldenLoad(name string) (data []byte) {
	golden := filepath.Join("testdata", name)
	data, _ = ioutil.ReadFile(golden)
	return
}

func goldenEquals(t *testing.T, name string, actual []byte) {
	golden := filepath.Join("testdata", name+".golden")
	var expected []byte
	if *update {
		ioutil.WriteFile(golden, actual, 0644)
		expected = actual
	} else {
		expected, _ = ioutil.ReadFile(golden)
	}
	assert.Equal(t, actual, expected)
}
