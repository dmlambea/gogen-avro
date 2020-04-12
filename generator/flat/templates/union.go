package templates

const UnionTemplate = `
import (
	"fmt"
	"io"
	"reflect"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/setters"
)

type {{ .GoType }}Type int64

const (
{{- range $i, $child := .Children }}
	{{- if $.OptionalIndex | ne $i }}
	{{ $.GoType }}Type{{ $child.Name }} {{ $.GoType }}Type = {{ $i }}
	{{- end }}
{{- end }}
)

type {{ .Name }} struct {
	setters.BaseUnion
}

func (u *{{ .Name }}) Set(value interface{}) {
	switch t := value.(type) {
{{- range $i, $child := .Children }}
	{{- if $.OptionalIndex | ne $i }}
	case {{ $child.GoType }}:
		u.Type = int64({{ $i }})
	{{- end }}
{{- end }}
	default:
		panic(fmt.Sprintf("invalid union type %T for {{ .Name }}", t))
	}
	u.Value = value
}

func write{{ .Name }}(u {{ if .IsOptional }}*{{ end }}{{ .Name }}, w io.Writer) (err error) {
	{{- if .IsOptional }}
	if u == nil {
		return vm.WriteLong({{ .OptionalIndex }}, w)
	}{{- end }}
	if err = vm.WriteLong(u.Type, w); err != nil {
		return
	}

	switch {{ .GoType }}Type(u.Type) {
{{- range $i, $child := .Children }}
	{{- if $.OptionalIndex | ne $i }}
	case {{ $.GoType }}Type{{ $child.Name }}:
		err = {{ $child.SerializerMethod }}(u.Value, w)
	{{- end }}
{{- end }}
	default:
		panic("invalid union type for {{ .Name }}")
	}
	return
}

func (u {{ .Name }}) UnionTypes() []reflect.Type {
	return typesFor{{ .Name }}
}

var (
	typesFor{{ .Name }} = []reflect.Type{
{{- range $i, $child := .Children }}
	{{- if $.OptionalIndex | ne $i }}
		reflect.TypeOf((*{{$child.Type.GoType}})(nil)),
	{{- else }}
		reflect.TypeOf(nil),
	{{- end }}
{{- end }}	
	}
)
`
