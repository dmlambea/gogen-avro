package compiler

import (
	"fmt"

	"github.com/actgardner/gogen-avro/schema"
	"github.com/actgardner/gogen-avro/vm"
)

const (
	oDiscardable int = -1
	oSkippable   int = -2
)

func vmTypeFor(gt schema.GenericType) (t vm.Type, err error) {
	t = vm.TypeFromString(gt.Name())
	if t == vm.TypeError {
		err = fmt.Errorf("type %s is not a primitive type", gt.Name())
	}
	return
}

func asField(t schema.GenericType) *schema.FieldType {
	return t.(*schema.FieldType)
}

func assertNotReference(t schema.GenericType, which string) error {
	if _, ok := t.(*schema.Reference); ok {
		return fmt.Errorf("unresolved reference in %s: %s", which, t)
	}
	return nil
}

func isUnionType(t schema.GenericType) bool {
	_, ok := t.(*schema.UnionType)
	return ok
}

// refineOrder extacts all special field indexes (opDiscardable and opSkippable)
func refineOrder(src []int) (tgt []int) {
	tgt = make([]int, len(src))
	cur := 0
	for i := range src {
		if src[i] >= 0 {
			tgt[cur] = src[i]
			cur++
		}
	}
	return tgt[:cur]
}

// getReadOrder computes in which order the reader needs to read writer's output.
// As the algorithm resolves, it tries to detect if all valid fields come in ascending
// order. It that is the case, the compiler would optimize field rearraging.
func getReadOrder(wrt, rdr *schema.RecordType) (order []int, allAsc bool, err error) {
	lastIdx := -1
	allAsc = true // Ascending order is kept true until a non-natural ordering is detected

	// For every writer's field, let's see in what position the reader expects the value
	for _, wrtChild := range wrt.Children() {
		wrtFld := wrtChild.(*schema.FieldType)
		var rdrFld *schema.FieldType
		code := oDiscardable
		if rdr != nil {
			// All children within a record are fields, and must be found by name
			if rdrFld = rdr.FindFieldByNameOrAlias(wrtFld); rdrFld != nil {
				if !wrtFld.Type().IsReadableBy(rdrFld.Type(), make(schema.VisitMap)) {
					err = fmt.Errorf("incompatible schemas: field %s in reader has incompatible type in writer field %s", rdrFld.Name(), wrtFld.Name())
					return
				}
				code = rdrFld.Index()
				if code < lastIdx {
					allAsc = false
				}
				lastIdx = code
			}
		}
		order = append(order, code)
	}

	// The rest of reader's fields must be skipped
	var extras []int
	if rdr != nil {
		for idx := range rdr.Children() {
			if !inArray(order, idx) {
				extras = append(extras, oSkippable)
			}
		}
	}

	if len(extras) > 0 {
		order = append(order, extras...)
	}
	return
}

func inArray(arr []int, val int) bool {
	for i := range arr {
		if arr[i] == val {
			return true
		}
	}
	return false
}
