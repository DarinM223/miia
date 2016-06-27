package graph

import "reflect"

// ForNode is a node that listens on the subnodes and
// passes up values as they are passed up from the subnodes
// into the body node.
type ForNode struct {
	id          int
	subnodes    []Node
	body        Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewForNode(id int, collection Node, body Node) *ForNode {
	forNode := &ForNode{
		id:          id,
		body:        body,
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
func (n *ForNode) IsLoop() bool                  { return true }

func (n *ForNode) Run() {
	passUpCount := 0

	isLoop := n.body.IsLoop() // true if the node's body contains a loop node
	msgReceived := false      // true if the for loop received the collection
	currNode := 0             // the index of the current node if the for loop is sequential

	for {
		msg := <-n.inChan

		// If sequential run the next node.
		if isLoop && msgReceived {
			currNode++
			if currNode < len(n.subnodes) {
				go n.subnodes[currNode].Run()
			}
		}

		if msg.Type == QuitMsg {
			// Send the quit message to all of the subnodes.
			for _, subnode := range n.subnodes {
				subnode.Chan() <- msg
			}
			n.Destroy()
			break
		} else if msg.PassUp {
			// Send the message to all of the parent nodes.
			for _, parent := range n.parentChans {
				parent <- msg
			}
			passUpCount++
			if passUpCount >= len(n.subnodes) {
				n.Destroy()
				break
			}
		} else if !msgReceived {
			// On receiving an array, allocate the subnodes
			arr := reflect.ValueOf(msg.Data)
			if arr.Kind() == reflect.Array {
				n.subnodes = make([]Node, arr.Len())
				for i := 0; i < arr.Len(); i++ {
					// TODO(DarinM223): need to clone the body node.
					n.subnodes[i] = n.body
					n.subnodes[i].addParentChan(n.id, n.inChan)

					n.subnodes[len(n.subnodes)-1].Chan() <- Msg{
						ValueMsg,
						false,
						arr.Index(i).Interface(),
					}
				}
			}

			msgReceived = true

			// Check if the body node is a loop node.
			// If it is, then run each subnode and listen for input sequentially,
			// otherwise run all of the subnodes in parallel and listen for all of them.
			if !isLoop {
				for _, node := range n.subnodes {
					go node.Run()
				}
			} else if len(n.subnodes) > 0 {
				go n.subnodes[currNode].Run()
			}
		}
	}
}

func (n *ForNode) AddChild(child Node)    { child.addParentChan(n.id, n.inChan) }
func (n *ForNode) RemoveChild(child Node) { child.removeParentChan(n.id) }

func (n *ForNode) Destroy() {
	for _, node := range n.subnodes {
		n.RemoveChild(node)
	}
}

func (n *ForNode) addParentChan(id int, parentChan chan Msg) { n.parentChans[id] = parentChan }
func (n *ForNode) removeParentChan(id int)                   { delete(n.parentChans, id) }
