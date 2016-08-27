package graph

import (
	"net/http"
	"testing"
)

func TestGotoNode(t *testing.T) {
	g := NewGlobals()
	parentChan1, parentChan2 := make(chan Msg, InChanSize), make(chan Msg, InChanSize)

	urlNode := NewValueNode(g, g.GenID(), "http://www.google.com")
	gotoNode := NewGotoNode(g, g.GenID(), urlNode)
	gotoNode.ParentChans()[2] = parentChan1
	gotoNode.ParentChans()[3] = parentChan2

	g.Run()

	if msg, ok := <-parentChan1; ok {
		if msg, ok := msg.(ValueMsg); ok {
			if _, ok := msg.Data.(*http.Response); !ok {
				t.Errorf("Message is not an HTTP response: got %v", msg.Data)
			}
		} else {
			t.Errorf("Message is not a Value message, got %v", msg)
		}
	} else {
		t.Errorf("Parent channel 1 closed")
	}

	if msg, ok := <-parentChan2; ok {
		if msg, ok := msg.(ValueMsg); ok {
			if _, ok := msg.Data.(*http.Response); !ok {
				t.Errorf("Message is not an HTTP response: got %v", msg.Data)
			}
		} else {
			t.Errorf("Message is not a Value message, got %v", msg)
		}
	} else {
		t.Errorf("Parent channel 2 closed")
	}
}

func TestGotoNodeErrsOnNonString(t *testing.T) {
	g := NewGlobals()
	parentChan := make(chan Msg, InChanSize)

	urlNode := NewValueNode(g, g.GenID(), 20)
	gotoNode := NewGotoNode(g, g.GenID(), urlNode)
	gotoNode.ParentChans()[1] = parentChan

	g.Run()

	if msg, ok := <-parentChan; ok {
		if _, ok := msg.(ErrMsg); !ok {
			t.Errorf("Message is not an error: got %v", msg)
		}
	} else {
		t.Errorf("Parent channel closed")
	}
}

func TestGotoNodeErrsOnNonHTTP(t *testing.T) {
	g := NewGlobals()
	parentChan := make(chan Msg, InChanSize)

	urlNode := NewValueNode(g, g.GenID(), "www.google.com")
	gotoNode := NewGotoNode(g, g.GenID(), urlNode)
	gotoNode.ParentChans()[1] = parentChan

	g.Run()

	if msg, ok := <-parentChan; ok {
		if _, ok := msg.(ErrMsg); !ok {
			t.Errorf("Message is not an error: got %v", msg)
		}
	} else {
		t.Errorf("Parent channel closed")
	}
}
