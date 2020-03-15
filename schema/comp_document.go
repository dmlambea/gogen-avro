package schema

// DocumentedType is implemented by any AVRO type able to have a documentary
// text associated to it, like fixed, enums and records.
type DocumentedType interface {
	Doc() string
	SetDoc(string)
}

var (
	// Ensure interface implementation
	_ DocumentedType = &documentComponent{}
)

type documentComponent struct {
	doc string
}

func (t *documentComponent) Doc() string {
	return t.doc
}

func (t *documentComponent) SetDoc(doc string) {
	t.doc = doc
}
