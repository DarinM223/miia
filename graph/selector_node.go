package graph

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

type SelectorType byte

const (
	SelectorClass SelectorType = '.' // A CSS class to retrieve.
	SelectorID                 = '#' // A CSS id to retrieve.
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

func NewSelectorNode(globals *Globals, gotoNode Node, selectors []Selector) *SelectorNode {
	id := globals.GenerateID()
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
func (n *SelectorNode) Dependencies() []Node          { return nil }
func (n *SelectorNode) Clone(g *Globals) Node         { return NewSelectorNode(g, n.gotoNode, n.selectors) }

func (n *SelectorNode) Run() {
	defer n.destroy()

	msg := <-n.inChan
	if msg.Type == QuitMsg {
		return
	}

	data := Msg{ErrMsg, n.id, true, -1, errors.New("Message received is not a HTTP response")}

	if resp, ok := msg.Data.(*http.Response); ok {
		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			data = Msg{ErrMsg, n.id, true, -1, err}
		} else {
			bindings := make(map[string]string)
			for _, selector := range n.selectors {
				bindings[selector.Name] = doc.Find(selector.Selector).First().Text()
			}
			data = Msg{ValueMsg, n.id, true, -1, bindings}
		}
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
}

func (n *SelectorNode) destroy() {
	if n.gotoNode != nil {
		delete(n.gotoNode.ParentChans(), n.id)
		n.gotoNode = nil
	}
}
