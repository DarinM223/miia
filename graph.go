package main

// Msg contains the data being sent/received between Nodes.
type Msg struct {
	// Type is the type of the message being sent.
	Type int
	// PassUp is true when completed data is being
	// sent backwards from the child to the parent.
	PassUp bool
	// Data is the data contained in the message being sent.
	Data interface{}
}

// Node is a contained "actor" that sends/receives messages.
type Node interface {
	// ID returns the ID of the specified node.
	ID() int
	// Run runs the node in a goroutine.
	Run()
	// InChan accesses the input channel of the node.
	// This channel is for sending messages to the node.
	InChan() chan Msg
	// OutChan accesses the output channel of the node.
	// This channel is for receiving messages from the node.
	OutChan() chan Msg
	// ParentChans is a map of parent ids to parent input channels
	// to send the completed data to after finishing.
	ParentChans() map[int]chan Msg
	// AddChild adds a child Node and listens for the child node's
	// passed up messages
	AddChild(child Node)
	// RemoveChild removes a child Node and listens for
	RemoveChild(child Node)

	// addParentChan adds a new parent input channel to the node.
	// Only meant to be used by AddChild.
	addParentChan(id int, parentChan chan Msg)
	// removeParentChan removes a parent input channel by its id.
	// Only meant to be used by RemoveChild.
	removeParentChan(id int)
}

type ForNode struct {
	id          int
	collection  Node
	subnodes    []Node
	inChan      chan Msg
	outChan     chan Msg
	parentChans map[int]chan Msg
}

func NewForNode(id int, collection Node, subnodes []Node) *ForNode {
	return &ForNode{
		id:          id,
		collection:  collection,
		subnodes:    subnodes,
		inChan:      make(chan Msg),
		outChan:     make(chan Msg),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *ForNode) ID() int                       { return n.id }
func (n *ForNode) InChan() chan Msg              { return n.inChan }
func (n *ForNode) OutChan() chan Msg             { return n.outChan }
func (n *ForNode) ParentChans() map[int]chan Msg { return n.parentChans }

func (n *ForNode) Run() {
	// TODO(DarinM223): listen @ Collection's out channel and
	// then distribute the data to the subnodes.
}

func (n *ForNode) AddChild(child Node) {
	n.subnodes = append(n.subnodes, child)
	child.addParentChan(n.id, n.inChan)
}

func (n *ForNode) RemoveChild(child Node) {
	childIdx := -1
	for i := 0; i < len(n.subnodes); i++ {
		if n.subnodes[i].ID() == child.ID() {
			childIdx = i
			break
		}
	}
	if childIdx != -1 {
		n.subnodes = append(n.subnodes[:childIdx], n.subnodes[childIdx+1:]...)
	}
	child.removeParentChan(n.id)
}

func (n *ForNode) addParentChan(id int, parentChan chan Msg) {
	n.parentChans[id] = parentChan
}

func (n *ForNode) removeParentChan(id int) {
	delete(n.parentChans, id)
}

type GotoNode struct {
	id          int
	next        Node
	selectors   []SelectorExpr
	inChan      chan Msg
	outChan     chan Msg
	parentChans map[int]chan Msg
}

func NewGotoNode(id int, next Node) *GotoNode {
	return &GotoNode{
		id:          id,
		next:        next,
		selectors:   []SelectorExpr{},
		inChan:      make(chan Msg),
		outChan:     make(chan Msg),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *GotoNode) ID() int                       { return n.id }
func (n *GotoNode) InChan() chan Msg              { return n.inChan }
func (n *GotoNode) OutChan() chan Msg             { return n.outChan }
func (n *GotoNode) ParentChans() map[int]chan Msg { return n.parentChans }

func (n *GotoNode) Run() {
	// TODO(DarinM223): listen for input channel and
	// for every message received, send an HTTP request
	// to get the page retrieve the data using the selectors
	// and then either send into the next node if there is a
	// next node or pass the data back up.
}

func (n *GotoNode) AddChild(child Node) {
	if n.next != nil {
		n.RemoveChild(n.next)
	}
	n.next = child
	child.addParentChan(n.id, n.inChan)
}

func (n *GotoNode) RemoveChild(child Node) {
	if child.ID() == n.next.ID() {
		n.next.removeParentChan(n.id)
		n.next = nil
	}
}

func (n *GotoNode) addParentChan(id int, parentChan chan Msg) {
	n.parentChans[id] = parentChan
}

func (n *GotoNode) removeParentChan(id int) {
	delete(n.parentChans, id)
}
