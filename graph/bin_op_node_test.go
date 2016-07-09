package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"reflect"
	"testing"
)

var binOpNodeTests = []struct {
	op             tokens.Token
	a, b, expected interface{}
}{
	{
		tokens.RangeToken,
		5, 10,
		Msg{ValueMsg, 2, true, -1, []int{5, 6, 7, 8, 9}},
	},
	{
		tokens.RangeToken,
		10, 5,
		Msg{ValueMsg, 2, true, -1, []int{10, 9, 8, 7, 6}},
	},
	{
		tokens.EqualsToken,
		2, 2,
		Msg{ValueMsg, 2, true, -1, true},
	},
	{
		tokens.EqualsToken,
		2, 3,
		Msg{ValueMsg, 2, true, -1, false},
	},
	{
		tokens.EqualsToken,
		"hello", "hello",
		Msg{ValueMsg, 2, true, -1, true},
	},
	{
		tokens.EqualsToken,
		"hello", "world",
		Msg{ValueMsg, 2, true, -1, false},
	},
	{
		tokens.OrToken,
		true, false,
		Msg{ValueMsg, 2, true, -1, true},
	},
	{
		tokens.OrToken,
		false, false,
		Msg{ValueMsg, 2, true, -1, false},
	},
	{
		tokens.AndToken,
		true, false,
		Msg{ValueMsg, 2, true, -1, false},
	},
	{
		tokens.AndToken,
		true, true,
		Msg{ValueMsg, 2, true, -1, true},
	},
}

func TestBinOpNode(t *testing.T) {
	for _, test := range binOpNodeTests {
		globals := NewGlobals()
		parentChan := make(chan Msg, InChanSize)
		aNode := NewValueNode(globals, test.a)
		bNode := NewValueNode(globals, test.b)
		binOpNode := NewBinOpNode(globals, test.op, aNode, bNode)
		binOpNode.ParentChans()[4] = parentChan

		globals.Run()

		if msg, ok := <-parentChan; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages for test: %v received: expected %v got %v", test, test.expected, msg)
			}
		}
	}
}
