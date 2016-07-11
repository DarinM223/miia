package graph

import (
	"reflect"
	"testing"
)

var ifNodeTests = []struct {
	pred     interface{}
	conseq   interface{}
	alt      interface{}
	expected Msg
}{
	{
		true, "Conseq", "Alt",
		NewValueMsg(3, true, "Conseq"),
	},
	{
		false, "Conseq", "Alt",
		NewValueMsg(3, true, "Alt"),
	},
	{
		1, "Conseq", "Alt",
		NewErrMsg(3, true, IfPredicateErr),
	},
}

func TestIfNode(t *testing.T) {
	for _, test := range ifNodeTests {
		globals := NewGlobals()
		parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)

		pred := NewValueNode(globals, test.pred)
		conseq := NewValueNode(globals, test.conseq)
		alt := NewValueNode(globals, test.alt)

		ifNode := NewIfNode(globals, pred, conseq, alt)
		ifNode.ParentChans()[50] = parentChan1
		ifNode.ParentChans()[51] = parentChan2

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
