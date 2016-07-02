package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"reflect"
	"testing"
)

func TestUnOp(t *testing.T) {
	parentChan := make(chan Msg, InChanSize)
	node := NewValueNode(0, true)
	unOpNode := NewUnOpNode(1, tokens.NotToken, node)
	unOpNode.ParentChans()[2] = parentChan

	go node.Run()
	go unOpNode.Run()

	expected := Msg{ValueMsg, unOpNode.ID(), true, false}

	if msg, ok := <-parentChan; ok {
		if !reflect.DeepEqual(msg, expected) {
			t.Errorf("Different messages: received: expected %v got %v", expected, msg)
		}
	}
}
