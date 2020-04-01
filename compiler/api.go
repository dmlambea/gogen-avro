// Compiler has methods to generate GADGT VM bytecode from Avro schemas
package compiler

import (
	"bytes"
	"fmt"

	"github.com/actgardner/gogen-avro/parser"
	"github.com/actgardner/gogen-avro/schema"
	"github.com/actgardner/gogen-avro/vm"
)

// Given two Avro schemas, compile them into a program which can read the data
// written by `writer` and store it in the structs generated for `reader`.
// If you're reading records from an OCF you can use the New<RecordType>Reader()
// method that's generated for you, which will parse the schemas automatically.
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
	if _, ok := reader.(*schema.RecordType); ok {
		// Make sure main is just a call to a record type
		if main.Size() != 1 || len(main.methodRefs) != 1 {
			err = fmt.Errorf("invalid program entry point for type %t: main size %d, nested methods %d",
				reader, main.Size(), len(main.methodRefs))
			return
		}
		// There should be a call instruction in position 0
		main = main.methodRefs[0]
		delete(c.methods, main.name)
	} else {
		main.append(vm.Ret())
	}
	linkedCode := c.link(main)
	return vm.NewProgram(linkedCode), nil
}

func parseSchema(s []byte) (schema.GenericType, error) {
	ns := parser.NewNamespace(false)
	return ns.ParseSchema(s)
}
