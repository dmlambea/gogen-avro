package templates

const EnumTemplate = `
import (
	"io"

	"github.com/actgardner/gogen-avro/vm"
)

{{ if ne .Doc "" -}}
// {{ .Doc }}
{{ end -}}
type {{ .Name }} int32

const (
{{- range $i, $symbol := .Symbols }}
	{{ $.Name }}{{ $symbol }} {{ $.Name }} = {{ $i }}
{{- end }}
)

func (e {{ .Name  }}) String() string {
	switch e {
{{- range $i, $s := .Symbols }}
	case {{ $.Name }}{{ $s }}:
		return "{{ $s }}"
{{- end }}
	default:
		return ""
	}
}

func (e {{ .Name  }}) FromString(symbol string) ({{ .Name }}, error) {
	switch symbol {
{{- range $i, $s := .Symbols }}
	case "{{ $s }}":
		return {{ $.Name }}{{ $s }}, nil
{{- end }}
	default:
		panic("invalid symbol for {{ .Name }}")
	}
}

func write{{ .Name }}(e {{ .Name }}, w io.Writer) (err error) {
	return vm.WriteInt(int32(e), w)
}

`
