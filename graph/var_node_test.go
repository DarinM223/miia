package graph

import (
	"github.com/DarinM223/miia/tokens"
	"reflect"
	"testing"
)

func TestVarNode(t *testing.T) {
	parentChan := make(chan Msg, 1)
	g := NewGlobals()
	varNode := NewVarNode(g, g.GenID(), "a")
	valueNode := NewValueNode(g, g.GenID(), 2)
	multOpNode := NewMultOpNode(g, g.GenID(), tokens.AddToken, []Node{varNode, valueNode})
	SetVarNodes(multOpNode, "a", 1)
	multOpNode.ParentChans()[20] = parentChan

	g.Run()

	expected := NewValueMsg(multOpNode.ID(), true, 3)

	if msg, ok := <-parentChan; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages received: expected %v got %v", expected, msg)
		}
	}
}
