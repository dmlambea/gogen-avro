package schema

// NamespacedType is implemented by any AVRO type able to have a namespace
// and a possible list of aliases associated to it, like fixed, enums and records.
type NamespacedType interface {
	QName() QName
	Aliases() []QName
	SetAliases([]QName)
}

var (
	// Ensure interface implementation
	_ NamespacedType = &namespaceComponent{}
)

type namespaceComponent struct {
	qname   QName
	aliases []QName
}

func (t *namespaceComponent) Name() string {
	return DefaultNamer.ToPublicName(t.qname.String())
}

func (t *namespaceComponent) setQName(qname QName) {
	t.qname = qname
}

func (t *namespaceComponent) QName() QName {
	return t.qname
}

func (t *namespaceComponent) Aliases() []QName {
	return t.aliases
}

func (t *namespaceComponent) SetAliases(aliases []QName) {
	t.aliases = aliases
}
