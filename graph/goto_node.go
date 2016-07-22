package graph

import (
	"errors"
	"net/http"
)

// GotoNode listens for a url value node
// and sends an http request to the url
// and passes up the response.
type GotoNode struct {
	id          int
	url         Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewGotoNode(globals *Globals, url Node) *GotoNode {
	id := globals.GenerateID()
	inChan := make(chan Msg, InChanSize)
	url.ParentChans()[id] = inChan

	gotoNode := &GotoNode{
		id:          id,
		url:         url,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, gotoNode)
	return gotoNode
}

func (n *GotoNode) ID() int                       { return n.id }
func (n *GotoNode) Chan() chan Msg                { return n.inChan }
func (n *GotoNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *GotoNode) Dependencies() []Node          { return []Node{n.url} }
func (n *GotoNode) Clone(g *Globals) Node         { return NewGotoNode(g, n.url.Clone(g)) }

func (n *GotoNode) Run() {
	defer destroyNode(n)

	msg, msgOk := (<-n.inChan).(*ValueMsg)

	var data Msg = NewErrMsg(n.id, true, errors.New("Message received is not a string"))
	if url, ok := msg.Data.(string); msgOk && ok {
		// Send an HTTP request to get and pass up the response.
		resp, err := http.Get(url)
		if err != nil {
			data = NewErrMsg(n.id, true, err)
		} else {
			data = NewValueMsg(n.id, true, resp)
		}
	}
	for _, parent := range n.parentChans {
		parent <- data
	}
}
