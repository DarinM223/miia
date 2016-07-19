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

type fanOutType interface {
	calcNumNodes() int
}

type varFanOut struct {
	// A variable number of subnodes.
	numSubnodes int
	// The for node for the variable.
	node *ForNode
}

type constFanOut struct {
	// A constant number of subnodes.
	numSubnodes int
}

type addFanOut struct{ a, b fanOutType }
type multFanOut struct{ a, b fanOutType }

func (t *varFanOut) calcNumNodes() int   { return t.numSubnodes }
func (t *constFanOut) calcNumNodes() int { return t.numSubnodes }
func (t *addFanOut) calcNumNodes() int   { return t.a.calcNumNodes() + t.b.calcNumNodes() }
func (t *multFanOut) calcNumNodes() int  { return t.a.calcNumNodes() * t.b.calcNumNodes() }

// setNodeFanOut is a helper function for
// SetNodesFanOut that returns the fan out expression
// for a certain node.
func setNodeFanOut(node Node, vars *[]*varFanOut) fanOutType {
	switch n := node.(type) {
	case *ForNode:
		v := &varFanOut{1, n}
		*vars = append(*vars, v)
		body := setNodeFanOut(n.body, vars)
		return &addFanOut{&multFanOut{v, body}, &constFanOut{1}}
	case *ValueNode, *VarNode:
		return &constFanOut{1}
	}

	var result fanOutType = nil
	for _, dep := range node.Dependencies() {
		fanOut := setNodeFanOut(dep, vars)
		if result == nil {
			result = fanOut
		} else {
			result = &addFanOut{result, fanOut}
		}
	}
	result = &addFanOut{result, &constFanOut{1}}
	return result
}

// SetNodesFanOut is a function that sets the fanout
// for all for nodes in the graph starting from the given node
// given the maximum limit of concurrently running goroutines.
// This function is to be run before running the nodes in the graph.
func SetNodesFanOut(node Node, totalNodes int) {
	var vars []*varFanOut
	fanOut := setNodeFanOut(node, &vars)

	if len(vars) < 1 {
		return
	}

	filled := make([]bool, len(vars))
	numFilled := 0

	index := len(vars) - 1
	for {
		if !filled[index] {
			v := vars[index]
			v.numSubnodes++
			if fanOut.calcNumNodes() > totalNodes {
				v.numSubnodes--
				filled[index] = true

				numFilled++
				if numFilled >= len(vars) {
					break
				}
			}
		}
		index++
		if index >= len(vars) {
			index = 0
		}
	}

	// Set the fan outs into the for nodes.
	for _, v := range vars {
		v.node.setFanOut(v.numSubnodes)
	}
}
