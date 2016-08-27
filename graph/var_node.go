package graph

import "errors"

// VarNode is a variable node that is set dynamically.
// A VarNode's message can be set either before it's running
// or while it's running. If the message is set before, the
// VarNode will just send the saved message.  If the message is
// set after, the VarNode will listen on its input channel and
// only send the saved message once it receives a value.
//
// In order to set a VarNode you would first call the setMsg method
// and then send a nil value to its input channel.
type VarNode struct {
	id          int
	name        string
	msg         Msg
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewVarNode(globals *Globals, id int, name string) *VarNode {
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
func (n *VarNode) Clone(g *Globals) Node {
	varNode := NewVarNode(g, g.GenID(), n.name)
	varNode.msg = n.msg
	return varNode
}

func (n *VarNode) run() Msg {
	if n.msg != nil {
		return n.msg.SetID(n.id)
	}

	if _, ok := <-n.inChan; !ok {
		return NewErrMsg(n.id, true, errors.New("Error receiving from channel"))
	}

	return n.msg.SetID(n.id)
}

// setMsg sets the message that the VarNode will send to its parents.
func (n *VarNode) setMsg(msg Msg) { n.msg = msg }
