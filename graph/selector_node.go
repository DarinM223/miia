package graph

import (
	"errors"
	"io/ioutil"
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

	data := Msg{ErrMsg, true, errors.New("Message received is not a HTTP response")}

	if resp, ok := msg.Data.(*http.Response); ok {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			data = Msg{ErrMsg, true, err}
		} else {
			bindings := make(map[string]interface{})
			var err error

			// Read body and use goquery to parse out selectors
			for _, selector := range n.selectors {
				err = handleSelector(bindings, selector, body)
				if err != nil {
					break
				}
			}

			if err != nil {
				data = Msg{ErrMsg, true, err}
			} else {
				data = Msg{ValueMsg, true, bindings}
			}
		}
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
}

func (n *SelectorNode) Destroy() {
	if n.gotoNode != nil {
		delete(n.gotoNode.ParentChans(), n.id)
		n.gotoNode = nil
	}
}

// handleSelector parses a selector from the body and stores the result in the data map
func handleSelector(data map[string]interface{}, selector Selector, body []byte) error {
	switch SelectorType(selector.Selector[0]) {
	case SelectorClass:
	case SelectorID:
	default:
		return errors.New("Invalid selector")
	}
	return nil
}
