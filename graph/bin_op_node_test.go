package graph

import (
	"github.com/DarinM223/miia/tokens"
	"reflect"
	"testing"
)

var binOpNodeTests = []struct {
	op             tokens.Token
	a, b, expected interface{}
}{
	{
		tokens.RangeToken,
		5, 10,
		NewValueMsg(2, true, []int{5, 6, 7, 8, 9}),
	},
	{
		tokens.RangeToken,
		10, 5,
		NewValueMsg(2, true, []int{10, 9, 8, 7, 6}),
	},
	{
		tokens.EqualsToken,
		2, 2,
		NewValueMsg(2, true, true),
	},
	{
		tokens.EqualsToken,
		2, 3,
		NewValueMsg(2, true, false),
	},
	{
		tokens.EqualsToken,
		"hello", "hello",
		NewValueMsg(2, true, true),
	},
	{
		tokens.EqualsToken,
		"hello", "world",
		NewValueMsg(2, true, false),
	},
	{
		tokens.OrToken,
		true, false,
		NewValueMsg(2, true, true),
	},
	{
		tokens.OrToken,
		false, false,
		NewValueMsg(2, true, false),
	},
	{
		tokens.AndToken,
		true, false,
		NewValueMsg(2, true, false),
	},
	{
		tokens.AndToken,
		true, true,
		NewValueMsg(2, true, true),
	},
}

func TestBinOpNode(t *testing.T) {
	for _, test := range binOpNodeTests {
		globals := NewGlobals()
		parentChan := make(chan Msg, InChanSize)
		aNode := NewValueNode(globals, globals.GenID(), test.a)
		bNode := NewValueNode(globals, globals.GenID(), test.b)
		binOpNode := NewBinOpNode(globals, globals.GenID(), test.op, aNode, bNode)
		binOpNode.ParentChans()[4] = parentChan

		globals.Run()

		if msg, ok := <-parentChan; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages for test: %v received: expected %v got %v", test, test.expected, msg)
			}
		}
	}
}
