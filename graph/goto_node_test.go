package graph

import (
	"net/http"
	"testing"
)

func TestGotoNode(t *testing.T) {
	parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)

	urlNode := NewValueNode(0, "http://www.google.com")
	gotoNode := NewGotoNode(1, urlNode)
	gotoNode.ParentChans()[2] = parentChan1
	gotoNode.ParentChans()[3] = parentChan2

	go urlNode.Run()
	go gotoNode.Run()

	if msg, ok := <-parentChan1; ok {
		if _, ok := msg.Data.(*http.Response); !ok {
			t.Errorf("Message is not an HTTP response: got %v", msg.Data)
		}
	} else {
		t.Errorf("Parent channel 1 closed")
	}

	if msg, ok := <-parentChan2; ok {
		if _, ok := msg.Data.(*http.Response); !ok {
			t.Errorf("Message is not an HTTP response: got %v", msg.Data)
		}
	} else {
		t.Errorf("Parent channel 2 closed")
	}
}

func TestGotoNodeErrsOnNonString(t *testing.T) {
	parentChan := make(chan Msg, InChanSize)

	urlNode := NewValueNode(0, 20)
	gotoNode := NewGotoNode(1, urlNode)
	gotoNode.ParentChans()[1] = parentChan

	go urlNode.Run()
	go gotoNode.Run()

	if msg, ok := <-parentChan; ok {
		if msg.Type != ErrMsg {
			t.Errorf("Message is not an error: got %v", msg.Data)
		}
	} else {
		t.Errorf("Parent channel closed")
	}
}

func TestGotoNodeErrsOnNonHTTP(t *testing.T) {
	parentChan := make(chan Msg, InChanSize)

	urlNode := NewValueNode(0, "www.google.com")
	gotoNode := NewGotoNode(1, urlNode)
	gotoNode.ParentChans()[1] = parentChan

	go urlNode.Run()
	go gotoNode.Run()

	if msg, ok := <-parentChan; ok {
		if msg.Type != ErrMsg {
			t.Errorf("Message is not an error: got %v", msg.Data)
		}
	} else {
		t.Errorf("Parent channel closed")
	}
}
