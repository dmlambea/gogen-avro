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

func (comp *documentComponent) Doc() string {
	return comp.doc
}

func (comp *documentComponent) SetDoc(doc string) {
	comp.doc = doc
}
