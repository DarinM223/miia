package graph

import "net/http"

type SelectorType int

const (
	SelectorClass SelectorType = iota // A CSS class to retrieve.
	SelectorID                        // A CSS id to retrieve.
)

// Selector is binding from a value node that
// outputs a css string like `#id`
// to the name of the key in the output map after
// parsing all of the selectors like { button: ... }
type Selector struct {
	Name     string
	Selector Node
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

func NewSelectorNode(id int, gotoNode Node, selectors []Selector) *SelectorNode {
	inChan := make(chan Msg, InChanSize)
	gotoNode.ParentChans()[id] = inChan

	return &SelectorNode{
		id:          id,
		selectors:   selectors,
		gotoNode:    gotoNode,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}
}

func (n *SelectorNode) ID() int                       { return n.id }
func (n *SelectorNode) Chan() chan Msg                { return n.inChan }
func (n *SelectorNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *SelectorNode) IsLoop() bool                  { return false }

func (n *SelectorNode) Run() {
	msg := <-n.inChan
	if msg.Type == QuitMsg {
		return
	}

	if resp, ok := msg.Data.(*http.Response); ok {
		_ = resp
		// TODO(DarinM223): read body and use goquery to parse out selectors
	}
}

func (n *SelectorNode) Destroy() {
	if n.gotoNode != nil {
		delete(n.gotoNode.ParentChans(), n.id)
		n.gotoNode = nil
	}
}
