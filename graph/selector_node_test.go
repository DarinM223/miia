package graph

import (
	"reflect"
	"testing"
)

var selectorNodeTests = []struct {
	url       string
	selectors []Selector
	expected  Msg
}{
	{
		"http://motherfuckingwebsite.com/",
		[]Selector{Selector{"why", "h2"}},
		Msg{ValueMsg, 2, true, map[string]string{"why": "Seriously, what the fuck else do you want?"}},
	},
}

func TestSelectorNode(t *testing.T) {
	for _, test := range selectorNodeTests {
		parentChan := make(chan Msg, InChanSize)
		stringNode := NewValueNode(0, test.url)
		gotoNode := NewGotoNode(1, stringNode)
		selectorNode := NewSelectorNode(2, gotoNode, test.selectors)
		selectorNode.ParentChans()[3] = parentChan

		go stringNode.Run()
		go gotoNode.Run()
		go selectorNode.Run()

		if msg, ok := <-parentChan; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel closed")
		}
	}
}
