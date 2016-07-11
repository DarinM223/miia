package graph

import (
	"reflect"
	"testing"
)

func TestValueNode(t *testing.T) {
	globals := NewGlobals()
	parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)

	node := NewValueNode(globals, 20)
	node.ParentChans()[1] = parentChan1
	node.ParentChans()[2] = parentChan2

	go node.Run()

	expected := NewValueMsg(0, true, 20)

	if msg, ok := <-parentChan1; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages received: expected %v got %v", expected, msg)
		}
	} else {
		t.Errorf("Parent channel 1 closed")
	}

	if msg, ok := <-parentChan2; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages received: expected %v got %v", expected, msg)
		}
	} else {
		t.Errorf("Parent channel 2 closed")
	}
}
