package templates

const MapTemplate = `
import (
	"io"

	"github.com/actgardner/gogen-avro/vm"
)

type {{ .Name }} map[string]{{ .Type.GoType }}

func New{{ .Name }}() {{ .Name }} {
	return make({{ .Name }})
}

func write{{ .Name }}(r {{ .Name }}, w io.Writer) (err error) {
	if err = vm.WriteLong(int64(len(r)), w); err != nil || len(r) == 0 {
		return
	}
	for key, val := range r {
		if err = vm.WriteString(key, w); err != nil {
			return err
		}
		if err = {{ .Type.SerializerMethod }}(val, w); err != nil {
			return err
		}
	}
	return vm.WriteLong(0, w)
}

`
