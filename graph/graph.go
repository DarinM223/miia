package graph

type MsgType int

const (
	QuitMsg MsgType = iota
	ValueMsg
	ErrMsg
)

// Msg contains the data being sent/received between Nodes.
type Msg struct {
	// Type is the type of the message being sent.
	Type MsgType
	// ID is the id of the node sending the message.
	ID int
	// PassUp is true when completed data is being
	// sent backwards from the child to the parent.
	PassUp bool
	// Data is the data contained in the message being sent.
	Data interface{}
}

// The size of the input channel buffers for nodes.
const InChanSize = 5

// Node is a contained "actor" that sends/receives messages.
type Node interface {
	// ID returns the ID of the specified node.
	ID() int
	// Run runs the node in a goroutine.
	Run()
	// Chan is the input channel for the node.
	Chan() chan Msg
	// ParentChans is a map of parent ids to parent input channels.
	ParentChans() map[int]chan Msg
	// Clone returns a cloned Node.
	Clone(*Globals) Node
	// Dependencies returns the dependency nodes for the node.
	Dependencies() []Node
}

// IsLoop returns true if a node is a loop node or
// depends on a loop node and false otherwise.
func IsLoop(node Node) bool {
	var queue []Node
	queue = append(queue, node)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		if _, ok := n.(*ForNode); ok {
			return true
		} else {
			for _, dep := range n.Dependencies() {
				queue = append(queue, dep)
			}
		}
	}
	return false
}

// SetVar traverses the graph setting variable nodes.
func SetVar(node Node, name string, value interface{}) {
	var queue []Node
	queue = append(queue, node)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		if varNode, ok := n.(*VarNode); ok && varNode.name == name {
			varNode.inChan <- Msg{ValueMsg, node.ID(), true, value}
		} else {
			for _, dep := range n.Dependencies() {
				queue = append(queue, dep)
			}
		}
	}
}
