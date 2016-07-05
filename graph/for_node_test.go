package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"testing"
)

func TestForNode(t *testing.T) {
	globals := NewGlobals()
	parentChan := make(chan Msg, 6)

	collectionNode := NewValueNode(globals, []interface{}{1, 2, 3, 4, 5, 6})
	varNode := NewVarNode(globals, "i")
	valueNode := NewValueNode(globals, 1)
	bodyNode := NewMultOpNode(globals, tokens.AddToken, []Node{varNode, valueNode})
	forNode := NewForNode(globals, "i", collectionNode, bodyNode)
	forNode.ParentChans()[5] = parentChan

	globals.Run()

	expectedValues := map[int]bool{2: true, 3: true, 4: true, 5: true, 6: true, 7: true}

	for len(expectedValues) > 0 {
		if msg, ok := <-parentChan; ok {
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
