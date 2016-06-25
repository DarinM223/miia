package graph

// ForNode is a node that listens on the subnodes and
// passes up values as they are passed up from the subnodes
// into the body node.
type ForNode struct {
	id          int
	subnodes    []Node
	insertIdx   int
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewForNode(id int, collection Node, body Node, numSubnodes int) *ForNode {
	subnodes := make([]Node, numSubnodes)
	for i := 0; i < numSubnodes; i++ {
		subnodes[i] = body
	}

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

func (n *ForNode) addParentChan(id int, parentChan chan Msg) { n.parentChans[id] = parentChan }
func (n *ForNode) removeParentChan(id int)                   { delete(n.parentChans, id) }
