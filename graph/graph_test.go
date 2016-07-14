package graph

import (
	"github.com/DarinM223/miia/tokens"
	"testing"
)

func TestContainsLoopNode(t *testing.T) {
	globals := NewGlobals()
	valueNodes := make([]Node, 2)
	valueNodes[0] = NewValueNode(globals, 2)
	valueNodes[1] = NewForNode(globals, "i", NewValueNode(globals, 3), NewValueNode(globals, 3))

	multOpNode := NewMultOpNode(globals, tokens.AddToken, valueNodes)
	if ContainsLoopNode(multOpNode) == false {
		t.Errorf("MultOp node with for loop node has ContainsLoopNode() return false instead of true")
	}

	valueNodes[1] = NewValueNode(globals, 3)

	multOpNode2 := NewMultOpNode(globals, tokens.AddToken, valueNodes)
	if ContainsLoopNode(multOpNode2) == true {
		t.Errorf("MultOp node without for loop node has ContainsLoopNode() return true instead of false")
	}
}
