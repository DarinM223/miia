package graph

import "sync"

type Globals struct {
	currID    int
	mutex     *sync.Mutex
	nodeMap   map[int]Node
	varValues map[int]interface{}
}

func NewGlobals() *Globals {
	return &Globals{
		currID:    0,
		mutex:     &sync.Mutex{},
		nodeMap:   make(map[int]Node),
		varValues: make(map[int]interface{}),
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
	n.nodeMap[id] = node
}

func (n *Globals) RegisterVar(id int, value interface{}) {
	n.varValues[id] = value
}

func (n *Globals) LookupVar(id int) (interface{}, bool) {
	value, ok := n.varValues[id]
	return value, ok
}

func (n *Globals) Run() {
	for _, node := range n.nodeMap {
		go node.Run()
	}
}
