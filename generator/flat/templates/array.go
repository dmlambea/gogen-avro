package templates

const ArrayTemplate = `
import (
	"io"

	"github.com/actgardner/gogen-avro/vm"
)

type {{ .Name }} {{ .GoType }}

func write{{ .Name }}(r {{ .Name }}, w io.Writer) (err error) {
	if err = vm.WriteLong(int64(len(r)), w); err != nil || len(r) == 0 {
		return
	}
	for _, elem := range r {
		if err = {{ .Type.SerializerMethod }}(elem, w); err != nil {
			return err
		}
	}
	return vm.WriteLong(0, w)
}
`
