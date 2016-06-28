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
	// Destroy cleans up the resources before killing a Node.
	Destroy()
	// IsLoop is true if the node or any of the subnodes is a loop node.
	IsLoop() bool
}
