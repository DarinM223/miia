package graph

import (
	"errors"

	"github.com/DarinM223/miia/tokens"
)

// UnOpNode listens for a value and applies
// an operator to the value when it is received.
type UnOpNode struct {
	id          int
	operator    tokens.Token
	inChan      chan Msg
	node        Node
	parentChans map[int]chan Msg
}

func NewUnOpNode(globals *Globals, id int, operator tokens.Token, node Node) *UnOpNode {
	inChan := make(chan Msg, 1)
	node.ParentChans()[id] = inChan

	unOpNode := &UnOpNode{
		id:          id,
		operator:    operator,
		inChan:      inChan,
		node:        node,
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, unOpNode)
	return unOpNode
}

func (n *UnOpNode) ID() int                       { return n.id }
func (n *UnOpNode) Chan() chan Msg                { return n.inChan }
func (n *UnOpNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *UnOpNode) Dependencies() []Node          { return []Node{n.node} }
func (n *UnOpNode) Clone(g *Globals) Node {
	return NewUnOpNode(g, g.GenID(), n.operator, n.node.Clone(g))
}

func (n *UnOpNode) run() Msg {
	defer destroyNode(n)

	var errMsg Msg = NewErrMsg(n.id, true, errors.New("invalid type for UnOp NotToken"))

	val, ok := (<-n.inChan).(ValueMsg)
	if !ok {
		return errMsg
	}

	switch n.operator {
	case tokens.NotToken:
		value, ok := val.Data.(bool)
		if !ok {
			return errMsg
		}

		return NewValueMsg(n.id, true, !value)
	default:
		return NewErrMsg(n.id, true, errors.New("invalid UnOp operator"))
	}
}
