package graph

import (
	"reflect"
)

type forNodeType interface {
	forNode()
}

type valueNodeType struct {
	currIdx       int
	finishedNodes int
}

type streamNodeType struct {
	currIdx          int
	len              int
	visitedNodes     map[int]bool
	startedFirstNode bool
}

func (t *valueNodeType) forNode()  {}
func (t *streamNodeType) forNode() {}

// ForNode is a node that listens on the subnodes and
// passes up values as they are passed up from the subnodes
// into the body node.
type ForNode struct {
	id             int
	nodeType       forNodeType
	subnodes       map[int]Node
	name           string
	collection     Node
	body           Node
	inChan         chan Msg
	collectionChan chan Msg
	parentChans    map[int]chan Msg
	globals        *Globals
	nodeToIdx      map[int]int
}

func NewForNode(globals *Globals, name string, collection Node, body Node) *ForNode {
	id := globals.GenerateID()
	collectionChan := make(chan Msg, InChanSize)
	// Listen for collection's result
	collection.ParentChans()[id] = collectionChan

	forNode := &ForNode{
		id:             id,
		nodeType:       nil,
		subnodes:       make(map[int]Node),
		name:           name,
		collection:     collection,
		body:           body,
		inChan:         nil,
		collectionChan: collectionChan,
		parentChans:    make(map[int]chan Msg),
		globals:        globals,
		nodeToIdx:      make(map[int]int),
	}
	globals.RegisterNode(id, forNode)
	return forNode
}

func (n *ForNode) ID() int                       { return n.id }
func (n *ForNode) Chan() chan Msg                { return n.inChan }
func (n *ForNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *ForNode) Dependencies() []Node          { return []Node{n.collection, n.body} }
func (n *ForNode) Clone(g *Globals) Node         { return NewForNode(g, n.name, n.collection.Clone(g), n.body) }

func (n *ForNode) handleValueMsg(isLoop bool, msg *ValueMsg) {
	if n.nodeType == nil {
		n.nodeType = &valueNodeType{0, 0}
	}

	// On receiving an array, allocate the subnodes
	arr := reflect.ValueOf(msg.Data)
	if arr.Kind() == reflect.Array || arr.Kind() == reflect.Slice {
		n.inChan = make(chan Msg, arr.Len())
		for i := 0; i < arr.Len(); i++ {
			n.subnodes[i] = n.body.Clone(n.globals)
			n.subnodes[i].ParentChans()[n.id] = n.inChan
			n.nodeToIdx[n.subnodes[i].ID()] = i
			SetVarNodes(n.subnodes[i], n.name, arr.Index(i).Interface())
		}
	}

	// Check if the body node is a loop node.
	// If it is, then run each subnode and listen for input sequentially,
	// otherwise run all of the subnodes in parallel and listen for all of them.
	if !isLoop {
		for _, node := range n.subnodes {
			startNode(n.globals, node)
		}
	} else {
		// Run first subnode.
		startNode(n.globals, n.subnodes[0])
	}
}

func (n *ForNode) handleStreamMsg(isLoop bool, msg *StreamMsg) {
	if n.inChan == nil {
		n.inChan = make(chan Msg, msg.Len)
	}
	if n.nodeType == nil {
		n.nodeType = &streamNodeType{-1, msg.Len, make(map[int]bool), false}
	}

	i := msg.Idx
	n.subnodes[i] = n.body.Clone(n.globals)
	n.subnodes[i].ParentChans()[n.id] = n.inChan
	n.nodeToIdx[n.subnodes[i].ID()] = i
	SetVarNodes(n.subnodes[i], n.name, msg.Data)

	// Start node if the body is not a loop,
	// or if the index is -1.
	if nodeType, ok := n.nodeType.(*streamNodeType); ok {
		if !isLoop || nodeType.currIdx == -1 {
			startNode(n.globals, n.subnodes[i])
			nodeType.currIdx = i
		}
	}
}

func (n *ForNode) handlePassUpMsg(isLoop bool, msg Msg) bool {
	finished := false
	var data Msg
	switch nodeType := n.nodeType.(type) {
	case *valueNodeType:
		nodeType.finishedNodes++
		if nodeType.finishedNodes >= len(n.subnodes) {
			finished = true
		} else if isLoop {
			nodeType.currIdx++
			startNode(n.globals, n.subnodes[nodeType.currIdx])
		}

		if valueMsg, ok := msg.(*ValueMsg); ok {
			data = NewStreamMsg(n.id, true, n.nodeToIdx[valueMsg.ID()], len(n.subnodes), valueMsg.Data)
		} else {
			panic("Message is not a value message")
		}
	case *streamNodeType:
		nodeType.visitedNodes[nodeType.currIdx] = true
		if valueMsg, ok := msg.(*ValueMsg); ok {
			if len(nodeType.visitedNodes) >= nodeType.len {
				finished = true
			} else if isLoop {
				// Find the next node by looking through the nodes ignoring
				// already visited nodes. If there are no nodes, set the current
				// index to -1.
				nextNodeIdx := -1
				for i, _ := range n.subnodes {
					if visited, ok := nodeType.visitedNodes[i]; !ok || !visited {
						nextNodeIdx = i
						break
					}
				}

				nodeType.currIdx = nextNodeIdx
				if nextNodeIdx != -1 {
					startNode(n.globals, n.subnodes[nextNodeIdx])
				}
			}
			data = NewStreamMsg(n.id, true, n.nodeToIdx[valueMsg.ID()], nodeType.len, valueMsg.Data)
		} else {
			panic("Message is not a value message")
		}
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
	return finished
}

func (n *ForNode) Run() {
	defer destroyNode(n)

	isLoop := ContainsLoopNode(n.body) // true if the node's body contains a loop node

	for {
		select {
		case msg := <-n.collectionChan:
			switch m := msg.(type) {
			case *ValueMsg:
				n.handleValueMsg(isLoop, m)
			case *StreamMsg:
				n.handleStreamMsg(isLoop, m)
			default:
				panic("Invalid message from collectionChan received")
			}
		case msg := <-n.inChan:
			if finished := n.handlePassUpMsg(isLoop, msg); finished {
				break
			}
		}
	}
}
