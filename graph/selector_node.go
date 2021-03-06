package graph

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

// Selector is binding from a css string like `#id`
// to the name of the key in the output map after
// parsing all of the selectors like { button: ... }
type Selector struct {
	Name     string
	Selector string
}

// SelectorNode is a node that receives HTTP Responses and parses out CSS selectors
// and passes the result back to the parents.
type SelectorNode struct {
	id          int
	selectors   []Selector
	gotoNode    Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewSelectorNode(globals *Globals, id int, gotoNode Node, selectors []Selector) *SelectorNode {
	inChan := make(chan Msg, InChanSize)
	gotoNode.ParentChans()[id] = inChan

	selectorNode := &SelectorNode{
		id:          id,
		selectors:   selectors,
		gotoNode:    gotoNode,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, selectorNode)
	return selectorNode
}

func (n *SelectorNode) ID() int                       { return n.id }
func (n *SelectorNode) Chan() chan Msg                { return n.inChan }
func (n *SelectorNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *SelectorNode) Dependencies() []Node          { return []Node{n.gotoNode} }
func (n *SelectorNode) Clone(g *Globals) Node {
	return NewSelectorNode(g, g.GenID(), n.gotoNode.Clone(g), n.selectors)
}

func (n *SelectorNode) run() (data Msg) {
	defer n.destroy()

	var errMsg Msg = NewErrMsg(n.id, true, errors.New("Message received is not a HTTP response"))

	msg, ok := (<-n.inChan).(ValueMsg)
	if !ok {
		return errMsg
	}

	resp, ok := msg.Data.(*http.Response)
	if !ok {
		return errMsg
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return NewErrMsg(n.id, true, err)
	}

	bindings := make(map[string]string)
	for _, selector := range n.selectors {
		bindings[selector.Name] = doc.Find(selector.Selector).First().Text()
	}

	return NewValueMsg(n.id, true, bindings)
}

func (n *SelectorNode) destroy() {
	if n.gotoNode != nil {
		delete(n.gotoNode.ParentChans(), n.id)
		n.gotoNode = nil
	}
}
