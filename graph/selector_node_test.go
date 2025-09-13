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
		[]Selector{{"why", "h2"}},
		NewValueMsg(2, true, map[string]string{"why": "Seriously, what the fuck else do you want?"}),
	},
}

func TestSelectorNode(t *testing.T) {
	for _, test := range selectorNodeTests {
		g := NewGlobals()

		parentChan := make(chan Msg, InChanSize)
		stringNode := NewValueNode(g, g.GenID(), test.url)
		gotoNode := NewGotoNode(g, g.GenID(), stringNode)
		selectorNode := NewSelectorNode(g, g.GenID(), gotoNode, test.selectors)
		selectorNode.ParentChans()[3] = parentChan

		g.Run()

		if msg, ok := <-parentChan; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel closed")
		}
	}
}
