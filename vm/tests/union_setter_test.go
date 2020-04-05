package tests

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/actgardner/gogen-avro/vm"
	"github.com/actgardner/gogen-avro/vm/generated"
	"github.com/actgardner/gogen-avro/vm/setters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unionStringIntNodeFixture struct {
	input    []byte
	expected generated.UnionStringIntNode
}

var (
	unionStringIntNodeReaderProgram = []vm.Instruction{
		vm.Load(),
		vm.JmpEq(0, 10),
		vm.Mov(vm.TypeAcc),
		vm.JmpEq(1, 3),
		vm.JmpEq(2, 4),
		vm.JmpEq(3, 5),
		vm.Halt(),
		vm.Mov(vm.TypeString),
		vm.Jmp(3),
		vm.Mov(vm.TypeInt),
		vm.Jmp(1),
		vm.Record(1),
		vm.Ret(),
		vm.Mov(vm.TypeString),
		vm.RecordEq(1, 1),
		vm.Ret(),
		vm.Mov(vm.TypeInt),
		vm.RecordEq(1, -2),
		vm.Ret(),
	}

	unionStringIntNodeFixtures = []unionStringIntNodeFixture{
		{input: []byte{0}, expected: generated.UnionStringIntNode{}},
		{input: []byte{2, 8, 'T', 'e', 's', 't'}, expected: generated.UnionStringIntNode{setters.BaseUnion{Type: 1, Value: "Test"}}},
		{input: []byte{4, 84}, expected: generated.UnionStringIntNode{setters.BaseUnion{Type: 2, Value: int32(42)}}},
		{input: []byte{6, 12, 'N', 'o', 'd', 'e', '-', '1', 2, 2, 0}, expected: generated.UnionStringIntNode{
			setters.BaseUnion{
				3,
				generated.Node{
					Name: "Node-1",
					Addr: &generated.Address{Id: 1},
				},
			}},
		},
	}
)

func TestUnion(t *testing.T) {
	p := vm.NewProgram(unionStringIntNodeReaderProgram)

	for i, f := range unionStringIntNodeFixtures {
		var obj generated.UnionStringIntNode

		objSetter, err := setters.NewSetterFor(&obj)
		require.Nil(t, err)
		require.NotNil(t, objSetter)

		buf := bytes.NewBuffer(f.input)
		engine := vm.NewEngine(p, objSetter)
		err = engine.Run(buf)
		require.Nil(t, err)

		assert.Equal(t, f.expected, obj, fmt.Sprintf("Union %d fails", i))
	}
}
