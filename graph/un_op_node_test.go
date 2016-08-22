package graph

import (
	"github.com/DarinM223/miia/tokens"
	"reflect"
	"testing"
)

func TestUnOp(t *testing.T) {
	g := NewGlobals()
	parentChan := make(chan Msg, InChanSize)
	node := NewValueNode(g, g.GenerateID(), true)
	unOpNode := NewUnOpNode(g, g.GenerateID(), tokens.NotToken, node)
	unOpNode.ParentChans()[2] = parentChan

	g.Run()

	expected := NewValueMsg(unOpNode.ID(), true, false)

	if msg, ok := <-parentChan; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages: received: expected %v got %v", expected, msg)
		}
	}
}
