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

func NewIfNode(globals *Globals, id int, pred Node, conseq Node, alt Node) *IfNode {
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
func (n *IfNode) Dependencies() []Node          { return []Node{n.pred, n.conseq, n.alt} }

func (n *IfNode) Clone(g *Globals) Node {
	clonedPred := n.pred.Clone(g)
	clonedConseq := n.conseq.Clone(g)
	clonedAlt := n.alt.Clone(g)

	return NewIfNode(g, g.GenID(), clonedPred, clonedConseq, clonedAlt)
}

func (n *IfNode) run() Msg {
	defer destroyNode(n)

	var data Msg = NewErrMsg(n.id, true, IfPredicateErr)

	msg, ok := (<-n.inChan).(ValueMsg)
	if !ok {
		return data
	}

	pred, ok := msg.Data.(bool)
	if !ok {
		return data
	}

	if pred {
		data = <-n.conseqChan
	} else {
		data = <-n.altChan
	}

	return data.SetID(n.id)
}
