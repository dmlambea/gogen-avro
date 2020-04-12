package templates

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/actgardner/gogen-avro/schema"
)

var NoTemplateForType = errors.New("No template exists for supplied type")

func Template(t interface{}) (string, error) {
	var template string
	switch t.(type) {
	case *schema.ArrayType:
		template = ArrayTemplate
	case *schema.MapType:
		template = MapTemplate
	case *schema.UnionType:
		template = UnionTemplate
	case *schema.EnumType:
		template = EnumTemplate
	case *schema.FixedType:
		template = FixedTemplate
	case *schema.RecordType:
		template = RecordTemplate
	default:
		return "", NoTemplateForType
	}
	return Evaluate(template, t)
}

func Evaluate(templateStr string, obj interface{}) (string, error) {
	buf := &bytes.Buffer{}
	t, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", err
	}

	err = t.Execute(buf, obj)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
