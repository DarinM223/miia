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
		Msg{ValueMsg, 3, true, []int{5, 6, 7, 8, 9}},
	},
	{
		tokens.RangeToken,
		10, 5,
		Msg{ValueMsg, 3, true, []int{10, 9, 8, 7, 6}},
	},
	{
		tokens.EqualsToken,
		2, 2,
		Msg{ValueMsg, 3, true, true},
	},
	{
		tokens.EqualsToken,
		2, 3,
		Msg{ValueMsg, 3, true, false},
	},
	{
		tokens.EqualsToken,
		"hello", "hello",
		Msg{ValueMsg, 3, true, true},
	},
	{
		tokens.EqualsToken,
		"hello", "world",
		Msg{ValueMsg, 3, true, false},
	},
	{
		tokens.OrToken,
		true, false,
		Msg{ValueMsg, 3, true, true},
	},
	{
		tokens.OrToken,
		false, false,
		Msg{ValueMsg, 3, true, false},
	},
	{
		tokens.AndToken,
		true, false,
		Msg{ValueMsg, 3, true, false},
	},
	{
		tokens.AndToken,
		true, true,
		Msg{ValueMsg, 3, true, true},
	},
}

func TestBinOpNode(t *testing.T) {
	for _, test := range binOpNodeTests {
		parentChan := make(chan Msg, InChanSize)
		aNode := NewValueNode(1, test.a)
		bNode := NewValueNode(2, test.b)
		binOpNode := NewBinOpNode(3, test.op, aNode, bNode)
		binOpNode.ParentChans()[4] = parentChan

		go aNode.Run()
		go bNode.Run()
		go binOpNode.Run()

		if msg, ok := <-parentChan; ok {
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("Different messages for test: %v received: expected %v got %v", test, test.expected, msg)
			}
		}
	}
}
