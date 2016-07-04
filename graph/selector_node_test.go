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
		globals := NewGlobals()

		parentChan := make(chan Msg, InChanSize)
		stringNode := NewValueNode(globals, test.url)
		gotoNode := NewGotoNode(globals, stringNode)
		selectorNode := NewSelectorNode(globals, gotoNode, test.selectors)
		selectorNode.ParentChans()[3] = parentChan

		globals.Run()

		if msg, ok := <-parentChan; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel closed")
		}
	}
}
