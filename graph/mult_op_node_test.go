package graph

import (
	"github.com/DarinM223/miia/tokens"
	"reflect"
	"testing"
)

var multOpNodeTests = []struct {
	op       tokens.Token
	values   []any
	expected any
}{
	{
		tokens.AddToken,
		[]any{1, 2, 3, 4, 5, 6},
		NewValueMsg(6, true, 21),
	},
	{
		tokens.MulToken,
		[]any{1, 2, 3, 4, 5, 6},
		NewValueMsg(6, true, 720),
	},
	{
		tokens.AddToken,
		[]any{"a", "b", "c", "d", "e"},
		NewValueMsg(5, true, "abcde"),
	},
	{
		tokens.SubToken,
		[]any{10, 2, 2, 1, 1},
		NewValueMsg(5, true, 4),
	},
	{
		tokens.DivToken,
		[]any{6, 3, 2},
		NewValueMsg(3, true, 1),
	},
	{
		tokens.ListToken,
		[]any{1, 2, 3},
		NewValueMsg(3, true, []any{1, 2, 3}),
	},
}

func TestMultOpNode(t *testing.T) {
	for _, test := range multOpNodeTests {
		g := NewGlobals()
		valuesNodes := make([]Node, len(test.values))
		for i, value := range test.values {
			valuesNodes[i] = NewValueNode(g, g.GenID(), value)
		}

		multOpNode := NewMultOpNode(g, g.GenID(), test.op, valuesNodes)

		parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)
		multOpNode.ParentChans()[len(test.values)+1] = parentChan1
		multOpNode.ParentChans()[len(test.values)+2] = parentChan2

		g.Run()

		if msg, ok := <-parentChan1; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel 1 closed")
		}

		if msg, ok := <-parentChan2; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel 2 closed")
		}
	}
}
