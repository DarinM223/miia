package graph

import (
	"net/http"
	"testing"
)

func TestGotoNode(t *testing.T) {
	globals := NewGlobals()
	parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)

	urlNode := NewValueNode(globals, "http://www.google.com")
	gotoNode := NewGotoNode(globals, urlNode)
	gotoNode.ParentChans()[2] = parentChan1
	gotoNode.ParentChans()[3] = parentChan2

	globals.Run()

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
	globals := NewGlobals()
	parentChan := make(chan Msg, InChanSize)

	urlNode := NewValueNode(globals, 20)
	gotoNode := NewGotoNode(globals, urlNode)
	gotoNode.ParentChans()[1] = parentChan

	globals.Run()

	if msg, ok := <-parentChan; ok {
		if msg.Type != ErrMsg {
			t.Errorf("Message is not an error: got %v", msg.Data)
		}
	} else {
		t.Errorf("Parent channel closed")
	}
}

func TestGotoNodeErrsOnNonHTTP(t *testing.T) {
	globals := NewGlobals()
	parentChan := make(chan Msg, InChanSize)

	urlNode := NewValueNode(globals, "www.google.com")
	gotoNode := NewGotoNode(globals, urlNode)
	gotoNode.ParentChans()[1] = parentChan

	globals.Run()

	if msg, ok := <-parentChan; ok {
		if msg.Type != ErrMsg {
			t.Errorf("Message is not an error: got %v", msg.Data)
		}
	} else {
		t.Errorf("Parent channel closed")
	}
}
