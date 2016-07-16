package graph

import (
	"testing"
)

func TestCollectNode(t *testing.T) {
	parentChan := make(chan Msg, 1)
	g := NewGlobals()
	streamNode := NewForNode(g, "i", NewValueNode(g, []interface{}{1, 2, 3}), NewVarNode(g, "i"))
	collectNode := NewCollectNode(g, streamNode)
	collectNode.ParentChans()[69] = parentChan

	g.Run()

	expected := testUtils.NewValueMsg(true, []interface{}{1, 2, 3})

	if msg, ok := <-parentChan; ok {
		if !testUtils.CompareTestMsgToRealMsg(expected, msg) {
			t.Errorf("Different messages received: expected %v got %v", expected, msg)
		}
	}
}
