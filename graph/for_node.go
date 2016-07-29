package graph

import (
	"fmt"
	"reflect"
)

// forNodeType is the current state of the for loop node.
type forNodeType interface {
	forNode()
}

// valueNodeType is the state when a for loop
// receives a value message in its input channel.
type valueNodeType struct {
	currIdx       int
	finishedNodes int
}

// streamNodeType is the state when a for loop
// receives a stream message in its input channel.
type streamNodeType struct {
	numCurrIdxs      int
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
	id              int
	fanout          int
	nodeType        forNodeType
	subnodes        map[int]Node
	name            string
	collection      Node
	body            Node
	inChan          chan Msg
	collectionChan  chan Msg
	parentChans     map[int]chan Msg
	globals         *Globals
	nodeToIdx       map[int]int
	numMsgsReceived int
	isLoop          bool
}

func NewForNode(globals *Globals, name string, collection Node, body Node) *ForNode {
	id := globals.GenerateID()
	collectionChan := make(chan Msg, InChanSize)
	// Listen for collection's result
	collection.ParentChans()[id] = collectionChan

	forNode := &ForNode{
		id:              id,
		fanout:          1,
		nodeType:        nil,
		subnodes:        make(map[int]Node),
		name:            name,
		collection:      collection,
		body:            body,
		inChan:          nil,
		collectionChan:  collectionChan,
		parentChans:     make(map[int]chan Msg),
		globals:         globals,
		nodeToIdx:       make(map[int]int),
		numMsgsReceived: 0,
		isLoop:          false,
	}
	globals.RegisterNode(id, forNode)
	return forNode
}

func (n *ForNode) ID() int                       { return n.id }
func (n *ForNode) Chan() chan Msg                { return n.inChan }
func (n *ForNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *ForNode) Dependencies() []Node          { return []Node{n.collection, n.body} }
func (n *ForNode) Clone(g *Globals) Node {
	forNode := NewForNode(g, n.name, n.collection.Clone(g), n.body.Clone(g))
	forNode.setFanOut(n.fanout)
	return forNode
}

func (n *ForNode) handleValueMsg(msg *ValueMsg) {
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
	if !n.isLoop {
		for _, node := range n.subnodes {
			startNode(n.globals, node)
		}
	} else {
		nodeType := n.nodeType.(*valueNodeType)
		// Run fanout number of nodes.
		for i := 0; i < n.fanout && i < len(n.subnodes); i++ {
			startNode(n.globals, n.subnodes[i])
			nodeType.currIdx++
		}
	}
}

func (n *ForNode) handleStreamMsg(msg *StreamMsg) {
	if n.inChan == nil {
		n.inChan = make(chan Msg, msg.Len.Len())
	}
	if n.nodeType == nil {
		n.nodeType = &streamNodeType{-1, msg.Len.Len(), make(map[int]bool), false}
	}

	i := msg.Idx.PopIndex()
	n.subnodes[i] = n.body.Clone(n.globals)
	n.subnodes[i].ParentChans()[n.id] = n.inChan
	n.nodeToIdx[n.subnodes[i].ID()] = i
	SetVarNodes(n.subnodes[i], n.name, msg.Data)

	// Start node if the body is not a loop,
	// or if there are less running nodes than the fanout.
	if nodeType, ok := n.nodeType.(*streamNodeType); ok {
		if !n.isLoop || nodeType.numCurrIdxs < n.fanout {
			nodeType.visitedNodes[i] = true
			startNode(n.globals, n.subnodes[i])
			nodeType.numCurrIdxs++
		}
	}
}

func (n *ForNode) incValueNode(nodeType *valueNodeType) bool {
	nodeType.finishedNodes++
	if nodeType.finishedNodes >= len(n.subnodes) {
		return true
	} else if n.isLoop && nodeType.currIdx < len(n.subnodes) {
		startNode(n.globals, n.subnodes[nodeType.currIdx])
		nodeType.currIdx++
	}
	return false
}

func (n *ForNode) incStreamNode(nodeType *streamNodeType) bool {
	nodeType.numCurrIdxs--
	if len(nodeType.visitedNodes) >= nodeType.len {
		return true
	} else {
		// Find the next node by looking through the saved nodes ignoring
		// already visited nodes. If there are no saved nodes, don't do anything.
		nextNodeIdx := -1
		for _, i := range n.nodeToIdx {
			if visited, ok := nodeType.visitedNodes[i]; !ok || !visited {
				nextNodeIdx = i
				break
			}
		}

		if nextNodeIdx != -1 {
			nodeType.numCurrIdxs++
			nodeType.visitedNodes[nextNodeIdx] = true
			startNode(n.globals, n.subnodes[nextNodeIdx])
		}
	}
	return false
}

func (n *ForNode) handlePassUpMsg(msg Msg) bool {
	finished := false
	var data Msg
	switch nodeType := n.nodeType.(type) {
	case *valueNodeType:
		// For a value type since all of the values are already saved
		// you can start another node right away.

		if valueMsg, ok := msg.(*ValueMsg); ok {
			finished = n.incValueNode(nodeType)
			data = NewStreamMsg(
				n.id,
				true,
				NewStreamIndex(n.nodeToIdx[valueMsg.ID()]),
				NewStreamIndex(len(n.subnodes)),
				valueMsg.Data,
			)
		} else if streamMsg, ok := msg.(*StreamMsg); ok {
			// TODO(DarinM223): increment node if all streams for a subnode
			// received.
			streamMsg.setID(n.id)
			streamMsg.Idx.AddIndex(n.nodeToIdx[streamMsg.ID()])
			streamMsg.Len.AddIndex(len(n.subnodes))
			data = streamMsg
		} else {
			panic(fmt.Sprintf("Message is not a value or stream message: %v instead: %v", msg, reflect.TypeOf(msg)))
		}
	case *streamNodeType:
		// For a stream type you can only start another node if that node
		// has already been saved but not started. Saved nodes are in nodeToIdx.

		if valueMsg, ok := msg.(*ValueMsg); ok {
			finished = n.incStreamNode(nodeType)
			data = NewStreamMsg(
				n.id,
				true,
				NewStreamIndex(n.nodeToIdx[valueMsg.ID()]),
				NewStreamIndex(nodeType.len),
				valueMsg.Data,
			)
		} else if streamMsg, ok := msg.(*StreamMsg); ok {
			// TODO(DarinM223): increment node if all streams for a subnode
			// received.
			streamMsg.setID(n.id)
			streamMsg.Idx.AddIndex(n.nodeToIdx[streamMsg.ID()])
			streamMsg.Len.AddIndex(nodeType.len)
			data = streamMsg
		} else {
			panic(fmt.Sprintf("Message is not a value or stream message: %v", msg))
		}
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
	return finished
}

func (n *ForNode) setFanOut(fanout int) {
	n.fanout = fanout
}

func (n *ForNode) Run() {
	defer destroyNode(n)

	n.isLoop = ContainsLoopNode(n.body) // true if the node's body contains a loop node

	for {
		select {
		case msg := <-n.collectionChan:
			switch m := msg.(type) {
			case *ValueMsg:
				n.handleValueMsg(m)
			case *StreamMsg:
				n.handleStreamMsg(m)
			default:
				panic(fmt.Sprintf("Invalid message from collectionChan received: %v", msg))
			}
		case msg := <-n.inChan:
			n.numMsgsReceived++
			if finished := n.handlePassUpMsg(msg); finished {
				break
			}
		}
	}
}
