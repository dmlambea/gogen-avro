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

func (comp *namespaceComponent) Name() string {
	return DefaultNamer.ToPublicName(comp.qname.String())
}

func (comp *namespaceComponent) setQName(qname QName) {
	comp.qname = qname
}

func (comp *namespaceComponent) QName() QName {
	return comp.qname
}

func (comp *namespaceComponent) Aliases() []QName {
	return comp.aliases
}

func (comp *namespaceComponent) SetAliases(aliases []QName) {
	comp.aliases = aliases
}
