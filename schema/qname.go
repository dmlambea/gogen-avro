package schema

import "fmt"

// QName is a qualified name (name and namespace pair).
type QName struct {
	Name      string
	Namespace string
}

func (qn QName) String() string {
	if qn.Namespace == "" {
		return qn.Name
	}
	return fmt.Sprintf("%s.%s", qn.Namespace, qn.Name)
}
