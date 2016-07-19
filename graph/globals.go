package graph

import "sync"

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

func (n *Globals) GenerateID() int {
	n.mutex.Lock()
	id := n.currID
	n.currID++
	n.mutex.Unlock()
	return id
}

func (n *Globals) RegisterNode(id int, node Node) {
	n.mutex.Lock()
	n.nodeMap[id] = node
	n.mutex.Unlock()
}

func (n *Globals) Run() {
	for _, node := range n.nodeMap {
		go node.Run()
	}
}
