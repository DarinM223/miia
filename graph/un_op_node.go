package graph

import (
	"errors"
	"github.com/DarinM223/http-scraper/tokens"
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

func NewUnOpNode(id int, operator tokens.Token, node Node) *UnOpNode {
	inChan := make(chan Msg, 1)
	node.ParentChans()[id] = inChan
	return &UnOpNode{
		id:          id,
		operator:    operator,
		inChan:      inChan,
		node:        node,
		parentChans: make(map[int]chan Msg),
	}
}

func (n *UnOpNode) ID() int                       { return n.id }
func (n *UnOpNode) Chan() chan Msg                { return n.inChan }
func (n *UnOpNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *UnOpNode) IsLoop() bool                  { return n.node.IsLoop() }

func (n *UnOpNode) Run() {
	val := <-n.inChan

	var msg Msg
	switch n.operator {
	case tokens.NotToken:
		if data, ok := val.Data.(bool); ok {
			msg = Msg{ValueMsg, n.id, true, !data}
		} else {
			msg = Msg{ErrMsg, n.id, true, errors.New("Invalid type for UnOp NotToken")}
		}
	default:
		msg = Msg{ErrMsg, n.id, true, errors.New("Invalid UnOp operator")}
	}

	for _, parent := range n.parentChans {
		parent <- msg
	}
	n.destroy()
}

func (n *UnOpNode) destroy() {
	delete(n.node.ParentChans(), n.id)
}
