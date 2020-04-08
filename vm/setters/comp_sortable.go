package setters

type sortableFieldsComponent struct {
	fieldCount int
	indexes    []int
	fields     []interface{}
}

func newSortableFieldsComponent(fields []interface{}) sortableFieldsComponent {
	c := sortableFieldsComponent{
		fieldCount: len(fields),
		indexes:    make([]int, len(fields)),
		fields:     fields,
	}
	c.initSortOrder()
	return c
}

// get returns the i-th field, respecting the sort order
func (c sortableFieldsComponent) get(i int) interface{} {
	return c.fields[c.indexes[i]]
}

// set puts the given value onto the i-th field, respecting the sort order
func (c sortableFieldsComponent) set(i int, value interface{}) {
	c.fields[c.indexes[i]] = value
}

func (c sortableFieldsComponent) initSortOrder() {
	for i := range c.indexes {
		c.indexes[i] = i
	}
}

// sort replaces the ordering of the fields within this setter. the length of positions
// cannot exceed the length of the fields array. The position indexes must be between 0
// and len(fields)-1. All fields not referred to in the positions array are put in order
// of appearance at the end of the list.
func (c sortableFieldsComponent) sort(positions []int) {
	visited := make([]bool, c.fieldCount)
	for i := range positions {
		c.indexes[i] = positions[i]
		visited[positions[i]] = true
	}
	posCount := len(positions)
	if posCount < c.fieldCount {
		for i := 0; posCount < c.fieldCount; i++ {
			if !visited[i] {
				c.indexes[posCount] = i
				posCount++
			}
		}
	}
}
