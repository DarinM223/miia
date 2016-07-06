package graph

import "reflect"

// ForNode is a node that listens on the subnodes and
// passes up values as they are passed up from the subnodes
// into the body node.
type ForNode struct {
	id             int
	subnodes       []Node
	name           string
	collection     Node
	body           Node
	inChan         chan Msg
	collectionChan chan Msg
	parentChans    map[int]chan Msg
	globals        *Globals
}

func NewForNode(globals *Globals, name string, collection Node, body Node) *ForNode {
	id := globals.GenerateID()
	collectionChan := make(chan Msg, InChanSize)
	// Listen for collection's result
	collection.ParentChans()[id] = collectionChan

	forNode := &ForNode{
		id:             id,
		name:           name,
		collection:     collection,
		body:           body,
		inChan:         nil,
		collectionChan: collectionChan,
		parentChans:    make(map[int]chan Msg),
		globals:        globals,
	}
	globals.RegisterNode(id, forNode)
	return forNode
}

func (n *ForNode) ID() int                       { return n.id }
func (n *ForNode) Chan() chan Msg                { return n.inChan }
func (n *ForNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *ForNode) Dependencies() []Node          { return []Node{n.collection, n.body} }
func (n *ForNode) Clone(g *Globals) Node         { return NewForNode(g, n.name, n.collection.Clone(g), n.body) }

func (n *ForNode) Run() {
	defer destroyNode(n)

	isLoop := ContainsLoopNode(n.body) // true if the node's body contains a loop node
	currNode := 0                      // the index of the current node if the for loop is sequential
	collectionMsg := <-n.collectionChan

	// TODO(DarinM223): check if message type is a value or a stream.

	// On receiving an array, allocate the subnodes
	arr := reflect.ValueOf(collectionMsg.Data)
	if arr.Kind() == reflect.Array || arr.Kind() == reflect.Slice {
		n.subnodes = make([]Node, arr.Len())
		n.inChan = make(chan Msg, arr.Len())
		for i := 0; i < arr.Len(); i++ {
			n.subnodes[i] = n.body.Clone(n.globals)
			n.subnodes[i].ParentChans()[n.id] = n.inChan
			SetVarNodes(n.subnodes[i], n.name, arr.Index(i).Interface())
		}
	} else {
		panic("Invalid array type")
	}

	// Check if the body node is a loop node.
	// If it is, then run each subnode and listen for input sequentially,
	// otherwise run all of the subnodes in parallel and listen for all of them.
	if !isLoop {
		for _, node := range n.subnodes {
			startNode(n.globals, node)
		}
	}

	for passUpCount := 0; passUpCount < len(n.subnodes); passUpCount++ {
		msg := <-n.inChan

		// If sequential run the next node.
		if isLoop {
			currNode++
			if currNode < len(n.subnodes) {
				startNode(n.globals, n.subnodes[currNode])
			}
		}

		for _, parent := range n.parentChans {
			parent <- Msg{StreamMsg, n.id, true, msg.Data}
		}
	}
}
