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
	globals     *Globals
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewGotoNode(globals *Globals, id int, url Node) *GotoNode {
	inChan := make(chan Msg, InChanSize)
	url.ParentChans()[id] = inChan

	gotoNode := &GotoNode{
		id:          id,
		url:         url,
		globals:     globals,
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
func (n *GotoNode) Clone(g *Globals) Node         { return NewGotoNode(g, g.GenerateID(), n.url.Clone(g)) }

func (n *GotoNode) run() Msg {
	defer destroyNode(n)

	var errMsg Msg = NewErrMsg(n.id, true, errors.New("Message received is not a string"))

	msg, ok := (<-n.inChan).(ValueMsg)
	if !ok {
		return errMsg
	}

	url, ok := msg.Data.(string)
	if !ok {
		return errMsg
	}

	// Send an HTTP request to get and pass up the response.
	n.globals.RateLimit(url)
	resp, err := http.Get(url)
	if err != nil {
		return NewErrMsg(n.id, true, err)
	}

	return NewValueMsg(n.id, true, resp)
}
