package graph

import (
	"fmt"
	"reflect"
)

// forNodeType is the current state of the for loop node.
type forNodeType interface {
	// increment is called when the passed up value for a node is received and
	// a new node is needed to replace the just finished node. It checks if all nodes
	// have finished and if not it finds and starts up the next node to run. It returns
	// a boolean that is true when all nodes have finished and false otherwise.
	increment(n *ForNode) bool
	// handleMsg is called when a the for node receives a passup message.
	// It handles the passed up message differently depending on the current state
	// and returns the message that the for node will pass up and a boolean that
	// is true when the for node is finished and false otherwise.
	handleMsg(n *ForNode, msg Msg) (Msg, bool)
}

// valueNodeType is the state when a for loop
// receives a value message in its input channel.
type valueNodeType struct {
	currIdx       int
	finishedNodes int
}

func (t *valueNodeType) increment(n *ForNode) bool {
	t.finishedNodes++
	if t.finishedNodes >= len(n.subnodes) {
		return true
	} else if n.isLoop && t.currIdx < len(n.subnodes) {
		currIdx := fmt.Sprintf("%d", t.currIdx)
		startNode(n.globals, n.subnodes[currIdx])
		t.currIdx++
	}
	return false
}

func (t *valueNodeType) handleMsg(n *ForNode, msg Msg) (Msg, bool) {
	// For a value type since all of the values are already saved
	// you can start another node right away.

	finished := t.increment(n)
	if valueMsg, ok := msg.(ValueMsg); ok {
		return NewStreamMsg(
			n.id,
			true,
			n.nodeToIdx[valueMsg.ID()],
			NewStreamIndex(len(n.subnodes)),
			valueMsg.Data,
		), finished
	} else if streamMsg, ok := msg.(StreamMsg); ok {
		streamMsg.Idx = streamMsg.Idx.Append(n.nodeToIdx[streamMsg.ID()])
		streamMsg.Len = streamMsg.Len.AddIndex(len(n.subnodes))
		return streamMsg.SetID(n.id), finished
	}

	err := fmt.Errorf("message is not a value or stream message: %v instead: %v", msg, reflect.TypeOf(msg))
	return NewErrMsg(n.id, true, err), true
}

// streamNodeType is the state when a for loop
// receives a stream message in its input channel.
type streamNodeType struct {
	numCurrIdxs      int
	len              StreamIndex
	visitedNodes     map[string]bool
	startedFirstNode bool
}

func (t *streamNodeType) increment(n *ForNode) bool {
	t.numCurrIdxs--
	if len(t.visitedNodes) >= t.len.Len() {
		return true
	}

	// Find the next node by looking through the saved nodes ignoring
	// already visited nodes. If there are no saved nodes, don't do anything.
	nextNodeIdx := ""
	for _, i := range n.nodeToIdx {
		if visited, ok := t.visitedNodes[i.String()]; !ok || !visited {
			nextNodeIdx = i.String()
			break
		}
	}

	if nextNodeIdx != "" {
		t.numCurrIdxs++
		t.visitedNodes[nextNodeIdx] = true
		startNode(n.globals, n.subnodes[nextNodeIdx])
	}
	return false
}

func (t *streamNodeType) handleMsg(n *ForNode, msg Msg) (Msg, bool) {
	// For a stream type you can only start another node if that node
	// has already been saved but not started. Saved nodes are in nodeToIdx.

	finished := t.increment(n)
	if valueMsg, ok := msg.(ValueMsg); ok {
		return NewStreamMsg(
			n.id,
			true,
			n.nodeToIdx[valueMsg.ID()],
			t.len,
			valueMsg.Data,
		), finished
	} else if streamMsg, ok := msg.(StreamMsg); ok {
		streamMsg.Idx = streamMsg.Idx.Append(n.nodeToIdx[streamMsg.ID()])
		streamMsg.Len = streamMsg.Len.Append(t.len)
		return streamMsg.SetID(n.id), finished
	}

	err := fmt.Errorf("message is not a value or stream message: %v", msg)
	return NewErrMsg(n.id, true, err), true
}

// ForNode is a node that listens to a collection and
// creates subnodes for every message received from the collection
// and sets the variable nodes for each subnode with the data in the message.
//
// For nodes always pass up data as stream messages and they can handle stream messages
// passed up in either the collection or the body.
//
// If the for node doesn't have nested for nodes in the body it runs all of the subnodes at once.
// Otherwise the number of nodes concurrently running at a time is determined by the fanout
// set from the SetNodesFanOut function in the same package.
type ForNode struct {
	id int
	// the number of nodes to run concurrently at a time.
	fanout int
	// the state of the for node (whether the collection is a value type or a stream type).
	nodeType       forNodeType
	subnodes       map[string]Node
	name           string
	collection     Node
	body           Node
	inChan         chan Msg
	collectionChan chan Msg
	parentChans    map[int]chan Msg
	globals        *Globals
	// mapping from node ID to the index of the collected array data.
	nodeToIdx map[int]StreamIndex
	// true if the for node contains a nested for node.
	isLoop bool
}

func NewForNode(globals *Globals, id int, name string, collection Node, body Node) *ForNode {
	collectionChan := make(chan Msg, InChanSize)
	// Listen for collection's result
	collection.ParentChans()[id] = collectionChan

	forNode := &ForNode{
		id:             id,
		fanout:         1,
		nodeType:       nil,
		subnodes:       make(map[string]Node),
		name:           name,
		collection:     collection,
		body:           body,
		inChan:         nil,
		collectionChan: collectionChan,
		parentChans:    make(map[int]chan Msg),
		globals:        globals,
		nodeToIdx:      make(map[int]StreamIndex),
		isLoop:         ContainsLoopNode(body),
	}
	globals.RegisterNode(id, forNode)
	return forNode
}

func (n *ForNode) ID() int                       { return n.id }
func (n *ForNode) Chan() chan Msg                { return n.inChan }
func (n *ForNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *ForNode) Dependencies() []Node          { return []Node{n.collection, n.body} }
func (n *ForNode) Clone(g *Globals) Node {
	forNode := NewForNode(g, g.GenID(), n.name, n.collection.Clone(g), n.body.Clone(g))
	forNode.setFanOut(n.fanout)
	return forNode
}

// handleValueMsg is called when the for node receives a value message
// from the collection channel.
func (n *ForNode) handleValueMsg(msg ValueMsg) {
	if n.nodeType == nil {
		n.nodeType = &valueNodeType{0, 0}
	}

	// On receiving an array, allocate the subnodes
	arr := reflect.ValueOf(msg.Data)
	if arr.Kind() == reflect.Array || arr.Kind() == reflect.Slice {
		n.inChan = make(chan Msg, arr.Len())
		for i := 0; i < arr.Len(); i++ {
			strIdx := fmt.Sprintf("%d", i)
			n.subnodes[strIdx] = n.body.Clone(n.globals)
			n.subnodes[strIdx].ParentChans()[n.id] = n.inChan
			n.nodeToIdx[n.subnodes[strIdx].ID()] = NewStreamIndex(i)
			SetVarNodes(n.subnodes[strIdx], n.name, arr.Index(i).Interface())
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
			strIdx := fmt.Sprintf("%d", i)
			startNode(n.globals, n.subnodes[strIdx])
			nodeType.currIdx++
		}
	}
}

// handleStreamMsg is called when the for node receives a stream message
// from the collection channel.
func (n *ForNode) handleStreamMsg(msg StreamMsg) {
	if n.inChan == nil {
		n.inChan = make(chan Msg, msg.Len.Len())
	}
	if n.nodeType == nil {
		n.nodeType = &streamNodeType{-1, msg.Len, make(map[string]bool), false}
	}

	i := msg.Idx.String()
	n.subnodes[i] = n.body.Clone(n.globals)
	n.subnodes[i].ParentChans()[n.id] = n.inChan
	n.nodeToIdx[n.subnodes[i].ID()] = msg.Idx
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

func (n *ForNode) handlePassUpMsg(msg Msg) bool {
	data, finished := n.nodeType.handleMsg(n, msg)
	BroadcastMsg(data, n.parentChans)
	return finished
}

func (n *ForNode) setFanOut(fanout int) {
	n.fanout = fanout
}

func (n *ForNode) run() Msg {
	defer destroyNode(n)

	for {
		select {
		case msg := <-n.collectionChan:
			switch m := msg.(type) {
			case ValueMsg:
				n.handleValueMsg(m)
			case StreamMsg:
				n.handleStreamMsg(m)
			default:
				panic(fmt.Sprintf("Invalid message from collectionChan received: %v", msg))
			}
		case msg := <-n.inChan:
			if finished := n.handlePassUpMsg(msg); finished {
				break
			}
		}
	}

	return nil
}
