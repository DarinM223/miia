package graph

import (
	"github.com/DarinM223/miia/tokens"
	"reflect"
	"testing"
)

func TestUnOp(t *testing.T) {
	globals := NewGlobals()
	parentChan := make(chan Msg, InChanSize)
	node := NewValueNode(globals, true)
	unOpNode := NewUnOpNode(globals, tokens.NotToken, node)
	unOpNode.ParentChans()[2] = parentChan

	globals.Run()

	expected := NewValueMsg(unOpNode.ID(), true, false)

	if msg, ok := <-parentChan; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages: received: expected %v got %v", expected, msg)
		}
	}
}
