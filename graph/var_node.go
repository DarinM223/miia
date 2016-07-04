package graph

type VarNode struct {
	id          int
	name        string
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewVarNode(globals *Globals, name string) *VarNode {
	id := globals.GenerateID()
	varNode := &VarNode{
		id:          id,
		name:        name,
		inChan:      make(chan Msg, 1),
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, varNode)
	return varNode
}

func (n *VarNode) ID() int                       { return n.id }
func (n *VarNode) Chan() chan Msg                { return n.inChan }
func (n *VarNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *VarNode) isLoop() bool                  { return false }
func (n *VarNode) setVar(name string, value interface{}) {
	if name == n.name {
		n.inChan <- Msg{ValueMsg, n.id, true, value}
	}
}

func (n *VarNode) Run() {
	msg := <-n.inChan
	if msg.Type == ValueMsg {
		for _, parent := range n.parentChans {
			parent <- msg
		}
	}
}

func (n *VarNode) Clone(globals *Globals) Node {
	varNode := NewVarNode(globals, n.name)
	varNode.parentChans = n.parentChans
	return varNode
}
