package setters

type eventFunc func(who Setter)

type exhaustNotifierComponent struct {
	fn eventFunc
}

func (n *exhaustNotifierComponent) setExhaustCallback(fn eventFunc) {
	n.fn = fn
}

func (n *exhaustNotifierComponent) hasExhaustCallback() bool {
	return n.fn != nil
}

func (n *exhaustNotifierComponent) trigger(who Setter) {
	n.fn(who)
}
