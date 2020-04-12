package templates

const RecordTemplate = `
import (
	"io"
	
	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/compiler"
)

{{ if ne .Doc "" -}}
// {{ .Doc }}
{{ end -}}
type {{ .Name }} struct { 
{{- range $field := .Children }}
	{{- if ne $field.Doc "" }}
	// {{ $field.Doc }}
	{{- end }}
	{{ $field.Name }} {{ $field.GoType }}
{{- end }}
}

func Deserialize{{ .Name }}(r io.Reader) (t {{ .Name }}, err error) {
	return Deserialize{{ .Name }}FromSchema(r, t.Schema())
}

func Deserialize{{ .Name }}FromSchema(r io.Reader, schema string) (t {{ .Name }}, err error) {
	p, err := compiler.CompileSchemaBytes([]byte(schema), []byte(t.Schema()))
	if err == nil {
		engine := vm.Engine{ Program: p	}
		err = engine.Run(r, &t)
	}
	return
}

func (r {{ .Name }}) Serialize(w io.Writer) error {
	return write{{ .Name }}(r, w)
}

func (r {{ .Name }}) Schema() string {
	return {{ printf "%q" .Schema }}
}

func write{{ .Name }}(r {{ .Name }}, w io.Writer) (err error) {
{{- range $f := .Children }}
	
	{{- if $f.IsSimple }}
	if r.{{ $f.Name }} == nil {
		if err = vm.WriteLong({{ $f.OptionalIndex }}, w); err != nil {
			return
		}
	} else {
		if err = vm.WriteLong({{ $f.NonOptionalIndex }}, w); err != nil {
			return
		}
		if err = {{ $f.SerializerMethod }}(*r.{{ $f.Name }}, w); err != nil {
			return
		}
	}
	{{- else }}
	if err = {{ $f.SerializerMethod }}(r.{{ $f.Name }}, w); err != nil {
		return
	}
	{{ end -}}
{{- end }}
	return
}
`
