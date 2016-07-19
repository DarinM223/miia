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

var setNodesFanOutTests = []struct {
	totalNodes, aFanout, bFanout int
}{
	{100, 5, 5},
	{25, 3, 2},
	{20, 2, 2},
	{10, 2, 1},
}

func TestSetNodesFanOut(t *testing.T) {
	for _, test := range setNodesFanOutTests {
		g := NewGlobals()

		// Graph:
		//                                   "x"
		//  "i"                 "x"          /
		// for B -> collect -> for A -> + ->
		//                                   \
		//                                    1
		forABody := NewMultOpNode(g, tokens.AddToken, []Node{NewVarNode(g, "x"), NewValueNode(g, 1)})
		forA := NewForNode(g, "x", NewValueNode(g, []interface{}{1, 2, 3, 4, 5}), forABody)
		forBBody := NewCollectNode(g, forA)
		forB := NewForNode(g, "i", NewValueNode(g, []interface{}{1, 2, 3, 4, 5}), forBBody)

		SetNodesFanOut(forB, test.totalNodes)

		if forB.fanout != test.bFanout {
			t.Errorf("Error for test %v: expected %v got %v", test, test.bFanout, forB.fanout)
		}

		if forA.fanout != test.aFanout {
			t.Errorf("Error for test %v: expected %v got %v", test, test.aFanout, forA.fanout)
		}
	}
}
