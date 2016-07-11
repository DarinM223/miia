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
func (n *VarNode) Dependencies() []Node          { return nil }
func (n *VarNode) Clone(globals *Globals) Node   { return NewVarNode(globals, n.name) }

func (n *VarNode) Run() {
	msg := <-n.inChan
	switch m := msg.(type) {
	case *ValueMsg:
		for _, parent := range n.parentChans {
			parent <- m
		}
	case *StreamMsg:
		panic("Stream message as var not implemented yet")
	default:
		panic("Unknown var message type")
	}
}
