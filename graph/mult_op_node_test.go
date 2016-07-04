package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"reflect"
	"testing"
)

var multOpNodeTests = []struct {
	op       tokens.Token
	values   []interface{}
	expected interface{}
}{
	{
		tokens.AddToken,
		[]interface{}{1, 2, 3, 4, 5, 6},
		Msg{ValueMsg, 6, true, 21},
	},
	{
		tokens.MulToken,
		[]interface{}{1, 2, 3, 4, 5, 6},
		Msg{ValueMsg, 6, true, 720},
	},
	{
		tokens.AddToken,
		[]interface{}{"a", "b", "c", "d", "e"},
		Msg{ValueMsg, 5, true, "abcde"},
	},
	{
		tokens.SubToken,
		[]interface{}{10, 2, 2, 1, 1},
		Msg{ValueMsg, 5, true, 4},
	},
	{
		tokens.DivToken,
		[]interface{}{6, 3, 2},
		Msg{ValueMsg, 3, true, 1},
	},
}

func TestMultOpNode(t *testing.T) {
	for _, test := range multOpNodeTests {
		globals := NewGlobals()
		valuesNodes := make([]Node, len(test.values))
		for i, value := range test.values {
			valuesNodes[i] = NewValueNode(globals, value)
		}

		multOpNode := NewMultOpNode(globals, test.op, valuesNodes)

		parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)
		multOpNode.ParentChans()[len(test.values)+1] = parentChan1
		multOpNode.ParentChans()[len(test.values)+2] = parentChan2

		globals.Run()

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

func TestMultOpNodeIsLoop(t *testing.T) {
	globals := NewGlobals()
	valueNodes := make([]Node, 2)
	valueNodes[0] = NewValueNode(globals, 2)
	valueNodes[1] = NewForNode(globals, NewValueNode(globals, 3), NewValueNode(globals, 3))

	multOpNode := NewMultOpNode(globals, tokens.AddToken, valueNodes)
	if multOpNode.isLoop() == false {
		t.Errorf("MultOp node with for loop node has isLoop() return false instead of true")
	}

	valueNodes[1] = NewValueNode(globals, 3)

	multOpNode2 := NewMultOpNode(globals, tokens.AddToken, valueNodes)
	if multOpNode2.isLoop() == true {
		t.Errorf("MultOp node without for loop node has isLoop() return true instead of false")
	}
}
