package graph

// The default size of the input channel buffers for nodes.
const InChanSize = 5

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
	// Clone returns a cloned Node.
	Clone(*Globals) Node
	// Dependencies returns the dependency nodes for the node.
	Dependencies() []Node
}

// ContainsLoopNode returns true if a node is a loop node or
// depends on a loop node and false otherwise.
func ContainsLoopNode(node Node) bool {
	var queue []Node
	queue = append(queue, node)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		if _, ok := n.(*ForNode); ok {
			return true
		} else {
			for _, dep := range n.Dependencies() {
				queue = append(queue, dep)
			}
		}
	}
	return false
}

// SetVarNodes traverses the graph setting variable nodes.
func SetVarNodes(node Node, name string, value interface{}) {
	var queue []Node
	queue = append(queue, node)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		if varNode, ok := n.(*VarNode); ok && varNode.name == name {
			varNode.inChan <- NewValueMsg(node.ID(), true, value)
		} else {
			for _, dep := range n.Dependencies() {
				queue = append(queue, dep)
			}
		}
	}
}

// startNode starts a node and its dependencies.
// Only to be used when a node is created dynamically
// and needs to be started after the other nodes.
func startNode(globals *Globals, node Node) {
	var queue []Node
	queue = append(queue, node)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		go n.Run()

		for _, dep := range n.Dependencies() {
			queue = append(queue, dep)
		}
	}
}

// destroyNode is called when a node is destroyed
// so that it can remove itself from its dependencies.
func destroyNode(node Node) {
	for _, dep := range node.Dependencies() {
		delete(dep.ParentChans(), node.ID())
	}
}
