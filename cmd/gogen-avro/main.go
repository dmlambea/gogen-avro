package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/actgardner/gogen-avro/generator"
	"github.com/actgardner/gogen-avro/generator/flat"
	"github.com/actgardner/gogen-avro/parser"
	"github.com/actgardner/gogen-avro/schema"
)

const fileComment = "// Code generated by github.com/actgardner/gogen-avro. DO NOT EDIT."

func main() {
	cfg := parseCmdLine()
	if len(cfg.files) == 0 {
		fmt.Fprintln(os.Stderr, "No valid input files specified")
		os.Exit(exitBadCommandLine)
	}

	var err error
	pkg := generator.NewPackage(cfg.packageName, codegenComment(cfg))
	namespace := parser.NewNamespace(cfg.shortUnions)
	gen := flat.NewFlatPackageGenerator(pkg, cfg.containers)

	switch cfg.namespacedNames {
	case nsShort:
		n := generator.NewNamespaceNamer(true)
		generator.SetNamer(n)
		schema.DefaultNamer = n
	case nsFull:
		n := generator.NewNamespaceNamer(false)
		generator.SetNamer(n)
		schema.DefaultNamer = n
	}

	for _, fileName := range cfg.files {
		schema, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %q - %v\n", fileName, err)
			os.Exit(exitErrorReadingFile)
		}

		_, err = namespace.ParseSchema(schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding schema for file %q - %v\n", fileName, err)
			os.Exit(exitErrorParsingSchema)
		}
	}

	for _, def := range namespace.Roots {
		err = gen.Add(def)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating code for schema - %v\n", err)
			os.Exit(exitErrorGeneratingCode)
		}
	}

	err = pkg.WriteFiles(cfg.targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing source files to directory %q - %v\n", cfg.targetDir, err)
		os.Exit(exitErrorWritingFile)
	}
}

// codegenComment generates a comment informing readers they are looking at
// generated code. If source avro files are provided they're included in the comment
func codegenComment(c config) string {
	if !c.sourcesComment {
		return fileComment
	}
	sourcesComment := `%s
/*
 * %s
 */`
	var sourceBlock []string
	if len(c.files) == 1 {
		sourceBlock = append(sourceBlock, "SOURCE:")
	} else {
		sourceBlock = append(sourceBlock, "SOURCES:")
	}

	for _, source := range c.files {
		_, fName := filepath.Split(source)
		sourceBlock = append(sourceBlock, fmt.Sprintf(" *     %s", fName))
	}
	return fmt.Sprintf(sourcesComment, fileComment, strings.Join(sourceBlock, "\n"))

}
