package graph

import (
	"fmt"
	"time"
)

type VarNode struct {
	id          int
	name        string
	msg         Msg
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewVarNode(globals *Globals, name string) *VarNode {
	id := globals.GenerateID()
	varNode := &VarNode{
		id:          id,
		name:        name,
		msg:         nil,
		inChan:      make(chan Msg, 1),
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, varNode)
	return varNode
}

func (n *VarNode) ID() int                       { return n.id }
func (n *VarNode) Chan() chan Msg                { return n.inChan }
func (n *VarNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *VarNode) Dependencies() []Node          { return nil }
func (n *VarNode) Clone(globals *Globals) Node {
	varNode := NewVarNode(globals, n.name)
	varNode.msg = n.msg
	return varNode
}

func (n *VarNode) Run() {
	if n.msg != nil {
		for _, parent := range n.parentChans {
			parent <- NewValueMsg(n.id, true, n.msg.GetData())
		}
		return
	}

	select {
	case <-n.inChan:
		switch m := n.msg.(type) {
		case *ValueMsg:
			for _, parent := range n.parentChans {
				parent <- NewValueMsg(n.id, true, m.Data)
			}
		case *StreamMsg:
			panic("Stream message as var not implemented yet")
		default:
			panic("Unknown var message type")
		}
	case <-time.After(5 * time.Second):
		panic(fmt.Sprintf("Variable %v timed out", n.name))
	}
}

func (n *VarNode) setMsg(msg Msg) { n.msg = msg }
