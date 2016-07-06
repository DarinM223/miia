package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"reflect"
	"testing"
)

func TestVarNode(t *testing.T) {
	parentChan := make(chan Msg, 1)
	globals := NewGlobals()
	varNode := NewVarNode(globals, "a")
	valueNode := NewValueNode(globals, 2)
	multOpNode := NewMultOpNode(globals, tokens.AddToken, []Node{varNode, valueNode})
	SetVarNodes(multOpNode, "a", 1)
	multOpNode.ParentChans()[20] = parentChan

	globals.Run()

	expected := Msg{ValueMsg, multOpNode.ID(), true, 3}

	if msg, ok := <-parentChan; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages received: expected %v got %v", expected, msg)
		}
	}
}
