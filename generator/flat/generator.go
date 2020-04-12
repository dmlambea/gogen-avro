package flat

import (
	"github.com/actgardner/gogen-avro/generator"
	"github.com/actgardner/gogen-avro/generator/flat/templates"
	"github.com/actgardner/gogen-avro/schema"
)

// FlatPackageGenerator emits a file per generated type, all in a single Go package without handling namespacing
type FlatPackageGenerator struct {
	containers bool
	files      *generator.Package
}

func NewFlatPackageGenerator(files *generator.Package, containers bool) *FlatPackageGenerator {
	return &FlatPackageGenerator{
		containers: containers,
		files:      files,
	}
}

type namedDefinition interface {
	Name() string
}

func (f *FlatPackageGenerator) Add(def namedDefinition) (err error) {
	var contents string

	// Ignore simple unions
	if u, ok := def.(*schema.UnionType); !ok || !u.IsSimple() {
		if contents, err = templates.Template(def); err == nil {
			// If there's a template for this definition, add it to the package
			filename := generator.ToSnake(def.Name()) + ".go"
			f.files.AddFile(filename, contents)
		} else {
			if err != templates.NoTemplateForType {
				return err
			}
		}
	}

	if r, ok := def.(*schema.RecordType); ok && f.containers {
		if err := f.addRecordContainer(r); err != nil {
			return err
		}
	}

	if ct, ok := def.(schema.CompositeType); ok {
		for _, child := range ct.Children() {
			// Avoid references
			if subCt, ok := child.(schema.ComplexType); ok {
				if err := f.Add(subCt); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (f *FlatPackageGenerator) addRecordContainer(def *schema.RecordType) error {
	containerFilename := generator.ToSnake(def.Name()) + "_container.go"
	file, err := templates.Evaluate(templates.RecordContainerTemplate, def)
	if err != nil {
		return err
	}
	f.files.AddFile(containerFilename, file)
	return nil
}
