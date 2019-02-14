package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/actgardner/gogen-avro/generator"
	"github.com/actgardner/gogen-avro/types"
)

func main() {
	cfg := parseCmdLine()

	var err error
	pkg := generator.NewPackage(cfg.packageName)
	namespace := types.NewNamespace(cfg.shortUnions)

	switch cfg.namespacedNames {
	case nsShort:
		generator.SetNamer(generator.NewNamespaceNamer(true))
	case nsFull:
		generator.SetNamer(generator.NewNamespaceNamer(false))
	}

	for _, fileName := range cfg.files {
		schema, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %q - %v\n", fileName, err)
			os.Exit(2)
		}

		_, err = namespace.TypeForSchema(schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding schema for file %q - %v\n", fileName, err)
			os.Exit(3)
		}
	}

	err = namespace.AddToPackage(pkg, codegenComment(cfg.files), cfg.containers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code for schema - %v\n", err)
		os.Exit(4)
	}

	err = pkg.WriteFiles(cfg.targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing source files to directory %q - %v\n", cfg.targetDir, err)
		os.Exit(4)
	}
}

// codegenComment generates a comment informing readers they are looking at
// generated code and lists the source avro files used to generate the code
//
// invariant: sources > 0
func codegenComment(sources []string) string {
	const fileComment = `// Code generated by github.com/actgardner/gogen-avro. DO NOT EDIT.
/*
 * %s
 */`
	var sourceBlock []string
	if len(sources) == 1 {
		sourceBlock = append(sourceBlock, "SOURCE:")
	} else {
		sourceBlock = append(sourceBlock, "SOURCES:")
	}

	for _, source := range sources {
		_, fName := filepath.Split(source)
		sourceBlock = append(sourceBlock, fmt.Sprintf(" *     %s", fName))
	}

	return fmt.Sprintf(fileComment, strings.Join(sourceBlock, "\n"))
}
