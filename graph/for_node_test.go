package graph

import (
	"github.com/DarinM223/miia/tokens"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

var forNodeTests = []struct {
	collection, body Node
	name             string
	expected         DataNode
}{
	{
		testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5, 6}),
		testUtils.NewMultOpNode(tokens.AddToken, []Node{
			testUtils.NewVarNode("i"),
			testUtils.NewValueNode(1),
		}),
		"i",
		testUtils.NewStreamDataArr(2, 3, 4, 5, 6, 7),
	},
	{
		testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5, 6}),
		testUtils.NewMultOpNode(tokens.AddToken, []Node{
			testUtils.NewValueNode(1),
			testUtils.NewVarNode("i"),
		}),
		"i",
		testUtils.NewStreamDataArr(2, 3, 4, 5, 6, 7),
	},
	{
		testUtils.NewForNode(
			"a",
			testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5, 6}),
			testUtils.NewVarNode("a"),
		),
		testUtils.NewMultOpNode(tokens.SubToken, []Node{
			testUtils.NewVarNode("i"),
			testUtils.NewValueNode(1),
		}),
		"i",
		testUtils.NewStreamDataArr(0, 1, 2, 3, 4, 5),
	},
	{
		testUtils.NewForNode(
			"i",
			testUtils.NewValueNode([]interface{}{1, 2, 3}),
			testUtils.NewCollectNode(
				testUtils.NewForNode(
					"x",
					testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5, 6}),
					testUtils.NewVarNode("x"),
				),
			),
		),
		testUtils.NewValueNode(1),
		"a",
		testUtils.NewStreamDataArr(1, 1, 1),
	},
	{
		testUtils.NewForNode(
			"i",
			testUtils.NewForNode(
				"a",
				testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5}),
				testUtils.NewVarNode("a"),
			),
			testUtils.NewCollectNode(
				testUtils.NewForNode(
					"x",
					testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5, 6}),
					testUtils.NewVarNode("x"),
				),
			),
		),
		testUtils.NewValueNode(1),
		"b",
		testUtils.NewStreamDataArr(1, 1, 1, 1, 1),
	},
}

func TestForNode(t *testing.T) {
	for _, test := range forNodeTests {
		globals := NewGlobals()
		parentChan := make(chan Msg, 6)

		collectionNode := testUtils.GenerateTestNode(globals, test.collection)
		bodyNode := testUtils.GenerateTestNode(globals, test.body)

		forNode := NewForNode(globals, test.name, collectionNode, bodyNode)
		SetNodesFanOut(forNode, 20)
		forNode.ParentChans()[5] = parentChan

		globals.Run()

		firstMsg, ok := <-parentChan
		if !ok {
			t.Errorf("Error receiving from channel")
		}

		firstStreamMsg, ok := firstMsg.(*StreamMsg)
		if !ok {
			t.Errorf("Expected Stream Message, got %v", firstMsg)
		}

		result := NewDataNode(firstStreamMsg.Len)
		result.Set(firstStreamMsg.Idx, firstStreamMsg.Data)

		for i := 0; i < firstStreamMsg.Len.Len()-1; i++ {
			msg, ok := <-parentChan
			if !ok {
				t.Errorf("Error receiving from channel")
			}

			streamMsg, ok := msg.(*StreamMsg)
			if !ok {
				t.Errorf("Expected Stream Message, got %v", msg)
			}

			result.Set(streamMsg.Idx, streamMsg.Data)
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Different values: expected %s got %s", spew.Sdump(test.expected), spew.Sdump(result))
		}
	}
}
