package graph

// ValueNode is a node that only passes up a value back to the parents.
type ValueNode struct {
	id          int
	value       interface{}
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewValueNode(globals *Globals, value interface{}) *ValueNode {
	id := globals.GenerateID()
	valueNode := &ValueNode{
		id:          id,
		value:       value,
		inChan:      make(chan Msg, InChanSize),
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, valueNode)
	return valueNode
}

func (n *ValueNode) ID() int                       { return n.id }
func (n *ValueNode) Chan() chan Msg                { return n.inChan }
func (n *ValueNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *ValueNode) Dependencies() []Node          { return nil }

func (n *ValueNode) Run() {
	data := Msg{ValueMsg, n.id, true, n.value}
	for _, parent := range n.parentChans {
		parent <- data
	}
}

func (n *ValueNode) Clone(globals *Globals) Node {
	return NewValueNode(globals, n.value)
}
