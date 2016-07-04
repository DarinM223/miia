package graph

import (
	"errors"
	"net/http"
)

type GotoNode struct {
	id          int
	url         Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewGotoNode(id int, url Node) *GotoNode {
	inChan := make(chan Msg, InChanSize)
	url.ParentChans()[id] = inChan
	return &GotoNode{
		id:          id,
		url:         url,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}
}

func (n *GotoNode) ID() int                       { return n.id }
func (n *GotoNode) Chan() chan Msg                { return n.inChan }
func (n *GotoNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *GotoNode) IsLoop() bool                  { return false }

func (n *GotoNode) Run() {
	msg := <-n.inChan
	if msg.Type == QuitMsg {
		return
	}

	data := Msg{ErrMsg, n.id, true, errors.New("Message received is not a string")}
	if url, ok := msg.Data.(string); ok {
		// Send an HTTP request to get and pass up the response.
		resp, err := http.Get(url)
		if err != nil {
			data = Msg{ErrMsg, n.id, true, err}
		} else {
			data = Msg{ValueMsg, n.id, true, resp}
		}
	}
	for _, parent := range n.parentChans {
		parent <- data
	}
}

func (n *GotoNode) Clone() Node {
	clonedURL := n.url.Clone()
	retNode := NewGotoNode(n.id, clonedURL)
	retNode.parentChans = n.parentChans
	return retNode
}
