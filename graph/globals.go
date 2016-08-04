package graph

import "sync"

// Globals contains all of the nodes so that
// they can all be run at the start of the program.
type Globals struct {
	currID  int
	mutex   *sync.Mutex
	nodeMap map[int]Node
}

func NewGlobals() *Globals {
	return &Globals{
		currID:  0,
		mutex:   &sync.Mutex{},
		nodeMap: make(map[int]Node),
	}
}

// GenerateID generates a new ID for a new node.
func (n *Globals) GenerateID() int {
	n.mutex.Lock()
	id := n.currID
	n.currID++
	n.mutex.Unlock()
	return id
}

// RegisterNode registers a node for an id.
func (n *Globals) RegisterNode(id int, node Node) {
	n.mutex.Lock()
	n.nodeMap[id] = node
	n.mutex.Unlock()
}

// Run runs all of the nodes in the map.
func (n *Globals) Run() {
	for _, node := range n.nodeMap {
		go RunNode(node)
	}
}
