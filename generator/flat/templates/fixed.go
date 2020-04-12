package templates

const FixedTemplate = `
import (
	"io"
)

type {{ .Name }} {{ .GoType }}

func write{{ .Name }}(f {{ .Name }}, w io.Writer) error {
	_, err := w.Write(f[:])
	return err
}

`
