package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
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
}

func TestForNode(t *testing.T) {

	for _, test := range forNodeTests {
		globals := NewGlobals()
		parentChan := make(chan Msg, 6)

		collectionNode := testUtils.GenerateTestNode(globals, test.collection)
		bodyNode := testUtils.GenerateTestNode(globals, test.body)

		forNode := NewForNode(globals, test.name, collectionNode, bodyNode)
		forNode.ParentChans()[5] = parentChan

		globals.Run()

		expectedValues := make(map[int]bool, len(test.expectedValues))
		for _, v := range test.expectedValues {
			expectedValues[v] = true
		}

		for len(expectedValues) > 0 {
			if msg, ok := <-parentChan; ok {
				if msg.Type != StreamMsg {
					t.Errorf("Expected Stream Message Type, got %d", msg.Type)
				}

				value := msg.Data.(int)

				if _, ok := expectedValues[value]; ok {
					delete(expectedValues, value)
				} else {
					t.Errorf("Received unexpected message %v", msg.Data)
					break
				}
			}
		}
	}
}
