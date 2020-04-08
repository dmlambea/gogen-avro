package compiler

import (
	"bytes"
	"fmt"

	"github.com/actgardner/gogen-avro/parser"
	"github.com/actgardner/gogen-avro/schema"
	"github.com/actgardner/gogen-avro/vm"
)

// CompileSchemaBytes creates a runnable program which can read the data
// written by `writer` and store it in the structs generated for `reader`.
func CompileSchemaBytes(writer, reader []byte) (p vm.Program, err error) {
	var writerType schema.GenericType
	if writerType, err = parseSchema(writer); err != nil {
		return
	}

	var readerType schema.GenericType
	switch bytes.Equal(writer, reader) {
	case true:
		readerType = writerType
	default:
		if readerType, err = parseSchema(reader); err != nil {
			return
		}
	}

	return Compile(writerType, readerType)
}

// Compile creates a runnable program for the two parsed Avro schemas, which can read the data
// written by `writer` and store it in the structs generated for `reader`.
func Compile(writer, reader schema.GenericType) (p vm.Program, err error) {
	c := newCompiler()
	main := newMethod("") // main is anonymous

	err = c.compileType(main, writer, reader)
	if err != nil {
		return
	}

	// Try and optimize the main method if it happens that it's just
	// a call to another one
	if main.Size() == 1 && main.code[0].IsRecordType() {
		// There should be a call instruction in position 0
		subroutine, ok := main.methodRefs[0]
		if !ok {
			err = fmt.Errorf("invalid program entry point for type %s: main size %d, nested methods %d",
				reader.Name(), main.Size(), len(main.methodRefs))
			return
		}
		main = subroutine
		delete(c.methods, subroutine.name)
	} else {
		// Subroutines have already a Ret instruction, but main don't
		main.append(vm.Ret())
	}
	linkedCode := c.link(main)
	return vm.NewProgram(linkedCode, c.errors), nil
}

func parseSchema(s []byte) (schema.GenericType, error) {
	ns := parser.NewNamespace(false)
	return ns.ParseSchema(s)
}
