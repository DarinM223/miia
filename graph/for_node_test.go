package graph

import (
	"github.com/DarinM223/miia/tokens"
	"testing"
)

var forNodeTests = []struct {
	collection, body Node
	name             string
	expectedValues   []int
}{
	{
		testUtils.NewValueNode([]interface{}{1, 2, 3, 4, 5, 6}),
		testUtils.NewMultOpNode(tokens.AddToken, []Node{
			testUtils.NewVarNode("i"),
			testUtils.NewValueNode(1),
		}),
		"i",
		[]int{2, 3, 4, 5, 6, 7},
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
		[]int{0, 1, 2, 3, 4, 5},
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
		[]int{1, 1, 1},
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
		[]int{1, 1, 1, 1, 1},
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

		expectedValues := make(map[int]int, len(test.expectedValues))
		for i, v := range test.expectedValues {
			expectedValues[v] = i
		}

		for len(expectedValues) > 0 {
			if msg, ok := <-parentChan; ok {
				if msg, ok := msg.(*StreamMsg); ok {
					value := msg.Data.(int)

					if msg.Len != len(test.expectedValues) {
						t.Errorf("Stream length different: expected %v got %v", len(test.expectedValues), msg.Len)
					}
					if _, ok := expectedValues[value]; ok && test.expectedValues[msg.Idx] == value {
						delete(expectedValues, value)
					} else {
						t.Errorf("Received unexpected message %v", msg.Data)
						break
					}
				} else {
					t.Errorf("Expected Stream Message, got %v", msg)
				}
			} else {
				t.Errorf("Error with channel")
			}
		}
	}
}
