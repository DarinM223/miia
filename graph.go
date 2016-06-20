package main

const (
	QuitMsg = iota
)

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
	// Chan is the input channel for the node.
	Chan() chan Msg
	// ParentChans is a map of parent ids to parent input channels.
	ParentChans() map[int]chan Msg
	// AddChild adds a child Node.
	AddChild(child Node)
	// RemoveChild removes a child Node.
	RemoveChild(child Node)
	// Destroy cleans up the resources before killing a Node.
	Destroy()

	// addParentChan adds a new parent input channel to the node.
	// Only meant to be used by AddChild.
	addParentChan(id int, parentChan chan Msg)
	// removeParentChan removes a parent input channel by its id.
	// Only meant to be used by RemoveChild.
	removeParentChan(id int)
}

type ForNode struct {
	id          int
	subnodes    []Node
	insertIdx   int
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewForNode(id int, collection Node, subnodes []Node) *ForNode {
	forNode := &ForNode{
		id:          id,
		insertIdx:   0,
		subnodes:    subnodes,
		inChan:      make(chan Msg),
		parentChans: make(map[int]chan Msg),
	}
	// Listen for collection's result
	collection.addParentChan(forNode.id, forNode.inChan)
	return forNode
}

func (n *ForNode) ID() int                       { return n.id }
func (n *ForNode) Chan() chan Msg                { return n.inChan }
func (n *ForNode) ParentChans() map[int]chan Msg { return n.parentChans }

func (n *ForNode) Run() {
	passUpCount := 0
	for {
		msg := <-n.inChan
		if msg.Type == QuitMsg {
			for _, subnode := range n.subnodes {
				subnode.Chan() <- msg
			}
			n.Destroy()
			break
		} else if msg.PassUp {
			for _, parent := range n.parentChans {
				parent <- msg
			}
			passUpCount++
			if passUpCount >= len(n.subnodes) {
				n.Destroy()
				break
			}
		} else if len(n.subnodes) > 0 {
			n.subnodes[n.insertIdx].Chan() <- msg
			n.insertIdx++
			if n.insertIdx >= len(n.subnodes) {
				n.insertIdx = 0
			}
		}
	}
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
		if n.insertIdx >= len(n.subnodes) {
			n.insertIdx = len(n.subnodes) - 1
		} else if childIdx < n.insertIdx {
			n.insertIdx--
		}
	}
	child.removeParentChan(n.id)
}

func (n *ForNode) Destroy() {
	// TODO(DarinM223): test that this works as expected.
	for _, node := range n.subnodes {
		n.RemoveChild(node)
	}
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
	parentChans map[int]chan Msg
}

func NewGotoNode(id int, next Node) *GotoNode {
	return &GotoNode{
		id:          id,
		next:        next,
		selectors:   []SelectorExpr{},
		inChan:      make(chan Msg),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *GotoNode) ID() int                       { return n.id }
func (n *GotoNode) Chan() chan Msg                { return n.inChan }
func (n *GotoNode) ParentChans() map[int]chan Msg { return n.parentChans }

func (n *GotoNode) Run() {
	for {
		msg := <-n.inChan
		if msg.Type == QuitMsg {
			if n.next != nil {
				n.next.Chan() <- msg
			}
			n.Destroy()
			break
		} else if msg.PassUp {
			for _, parent := range n.parentChans {
				parent <- msg
			}
			n.Destroy()
			break
		} else {
			var data Msg
			// TODO(DarinM223): send an HTTP request to get
			// the page and then either send to next node
			// or pass up.
			if n.next != nil {
				n.next.Chan() <- data
			} else {
				data.PassUp = true
				for _, parent := range n.parentChans {
					parent <- data
				}
				n.Destroy()
				break
			}
		}
	}
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

func (n *GotoNode) Destroy() {
	if n.next != nil {
		n.RemoveChild(n.next)
		n.next = nil
	}
}

func (n *GotoNode) addParentChan(id int, parentChan chan Msg) {
	n.parentChans[id] = parentChan
}

func (n *GotoNode) removeParentChan(id int) {
	delete(n.parentChans, id)
}
