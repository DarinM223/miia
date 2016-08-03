package graph

// CollectNode listens for messages
// from a node that outputs a stream
// and collects and sends out a value message.
type CollectNode struct {
	id          int
	node        Node
	inChan      chan Msg
	parentChans map[int]chan Msg
	results     DataNode
}

func NewCollectNode(globals *Globals, node Node) *CollectNode {
	id := globals.GenerateID()
	inChan := make(chan Msg, InChanSize)
	node.ParentChans()[id] = inChan
	collectNode := &CollectNode{
		id:          id,
		node:        node,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
		results:     nil,
	}
	globals.RegisterNode(id, collectNode)
	return collectNode
}

func (n *CollectNode) ID() int                       { return n.id }
func (n *CollectNode) Chan() chan Msg                { return n.inChan }
func (n *CollectNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *CollectNode) Dependencies() []Node          { return []Node{n.node} }

func (n *CollectNode) Run() {
	msg, ok := (<-n.inChan).(*StreamMsg)
	if !ok {
		panic("CollectNode not receiving a stream messsage")
	}

	var data Msg
	n.results = NewDataNode(msg.Len)
	n.results.Set(msg.Idx, msg.Data)

	// Listen for stream messages from node and
	// collect them into an array and then once all of the messages have
	// been received, send a value message with the collected data.
	for i := 0; i < msg.Len.Len()-1; i++ {
		streamMsg, ok := (<-n.inChan).(*StreamMsg)
		if !ok {
			panic("CollectNode not receiving a stream messsage")
		}
		n.results.Set(streamMsg.Idx, streamMsg.Data)
	}

	data = NewValueMsg(n.id, true, n.results.Data())

	for _, parentChan := range n.parentChans {
		parentChan <- data
	}
}

func (n *CollectNode) Clone(g *Globals) Node {
	return NewCollectNode(g, n.node.Clone(g))
}
