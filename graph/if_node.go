package graph

import "errors"

var IfPredicateErr error = errors.New("Predicate return is not a boolean value")

// IfNode is a node that listens to the predicate node,
// the consequence node (the node when predicate is true),
// and the alternate node (the node when predicate is false),
// and either returns the result of the consequence or the alternate node.
type IfNode struct {
	id          int
	pred        Node
	conseq      Node
	alt         Node
	inChan      chan Msg
	conseqChan  chan Msg
	altChan     chan Msg
	parentChans map[int]chan Msg
}

func NewIfNode(globals *Globals, pred Node, conseq Node, alt Node) *IfNode {
	id := globals.GenerateID()
	inChan := make(chan Msg, InChanSize)
	conseqChan := make(chan Msg, InChanSize)
	altChan := make(chan Msg, InChanSize)

	pred.ParentChans()[id] = inChan
	conseq.ParentChans()[id] = conseqChan
	alt.ParentChans()[id] = altChan

	ifNode := &IfNode{
		id:          id,
		pred:        pred,
		conseq:      conseq,
		alt:         alt,
		inChan:      inChan,
		conseqChan:  conseqChan,
		altChan:     altChan,
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, ifNode)
	return ifNode
}

func (n *IfNode) ID() int                       { return n.id }
func (n *IfNode) Chan() chan Msg                { return n.inChan }
func (n *IfNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *IfNode) isLoop() bool                  { return n.pred.isLoop() || n.conseq.isLoop() || n.alt.isLoop() }
func (n *IfNode) setVar(name string, value interface{}) {
	n.pred.setVar(name, value)
	n.conseq.setVar(name, value)
	n.alt.setVar(name, value)
}

func (n *IfNode) Run() {
	data := Msg{ErrMsg, n.id, true, IfPredicateErr}

	msg := <-n.inChan
	if pred, ok := msg.Data.(bool); ok {
		if pred {
			data = <-n.conseqChan
		} else {
			data = <-n.altChan
		}
		data.ID = n.id
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
}

func (n *IfNode) Clone(globals *Globals) Node {
	clonedPred := n.pred.Clone(globals)
	clonedConseq := n.conseq.Clone(globals)
	clonedAlt := n.alt.Clone(globals)

	retNode := NewIfNode(globals, clonedPred, clonedConseq, clonedAlt)
	retNode.parentChans = n.parentChans
	return retNode
}
