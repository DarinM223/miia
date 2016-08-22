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
		g := NewGlobals()
		parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)

		pred := NewValueNode(g, g.GenerateID(), test.pred)
		conseq := NewValueNode(g, g.GenerateID(), test.conseq)
		alt := NewValueNode(g, g.GenerateID(), test.alt)

		ifNode := NewIfNode(g, g.GenerateID(), pred, conseq, alt)
		ifNode.ParentChans()[50] = parentChan1
		ifNode.ParentChans()[51] = parentChan2

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
