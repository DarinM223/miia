package graph

import (
	"reflect"
	"testing"
)

var ifNodeTests = []struct {
	pred     Node
	conseq   Node
	alt      Node
	expected Msg
}{
	{
		NewValueNode(0, true),
		NewValueNode(1, "Conseq"),
		NewValueNode(2, "Alt"),
		Msg{ValueMsg, 69, true, "Conseq"},
	},
	{
		NewValueNode(0, false),
		NewValueNode(1, "Conseq"),
		NewValueNode(2, "Alt"),
		Msg{ValueMsg, 69, true, "Alt"},
	},
	{
		NewValueNode(0, 1),
		NewValueNode(1, "Conseq"),
		NewValueNode(2, "Alt"),
		Msg{ErrMsg, 69, true, IfPredicateErr},
	},
}

func TestIfNode(t *testing.T) {
	for _, test := range ifNodeTests {
		parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)
		ifNode := NewIfNode(69, test.pred, test.conseq, test.alt)
		ifNode.ParentChans()[50] = parentChan1
		ifNode.ParentChans()[51] = parentChan2

		go test.pred.Run()
		go test.conseq.Run()
		go test.alt.Run()
		go ifNode.Run()

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

func TestIfNodeIsLoop(t *testing.T) {
}
