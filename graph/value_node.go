package graph

// ValueNode is a node that only passes up a value back to the parents.
type ValueNode struct {
	id          int
	value       interface{}
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewValueNode(id int, value interface{}) *ValueNode {
	return &ValueNode{
		id:          id,
		value:       value,
		inChan:      make(chan Msg),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *ValueNode) ID() int                       { return n.id }
func (n *ValueNode) Chan() chan Msg                { return n.inChan }
func (n *ValueNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *ValueNode) AddChild(child Node)           { child.addParentChan(n.id, n.inChan) }
func (n *ValueNode) RemoveChild(child Node)        { child.removeParentChan(n.id) }
func (n *ValueNode) Destroy()                      {}

func (n *ValueNode) addParentChan(id int, parentChan chan Msg) { n.parentChans[id] = parentChan }
func (n *ValueNode) removeParentChan(id int)                   { delete(n.parentChans, id) }

func (n *ValueNode) Run() {
	data := Msg{ValueMsg, true, n.value}
	for _, parent := range n.parentChans {
		parent <- data
	}
}
