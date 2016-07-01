package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"reflect"
	"testing"
)

var multOpNodeTests = []struct {
	op       tokens.Token
	values   []interface{}
	expected interface{}
}{
	{
		tokens.AddToken,
		[]interface{}{1, 2, 3, 4, 5, 6},
		Msg{ValueMsg, true, 21},
	},
	{
		tokens.MulToken,
		[]interface{}{1, 2, 3, 4, 5, 6},
		Msg{ValueMsg, true, 720},
	},
}

func TestMultOpNode(t *testing.T) {
	for _, test := range multOpNodeTests {
		valuesNodes := make([]Node, len(test.values))
		for i, value := range test.values {
			valuesNodes[i] = NewValueNode(i, value)
		}

		multOpNode := NewMultOpNode(len(test.values), test.op, valuesNodes)

		parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)
		multOpNode.ParentChans()[len(test.values)+1] = parentChan1
		multOpNode.ParentChans()[len(test.values)+2] = parentChan2

		go multOpNode.Run()
		for _, valueNode := range valuesNodes {
			go valueNode.Run()
		}

		if msg, ok := <-parentChan1; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel 1 closed")
		}

		if msg, ok := <-parentChan2; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages received: expected %v got %v", test.expected, msg)
			}
		} else {
			t.Errorf("Parent channel 2 closed")
		}
	}
}
