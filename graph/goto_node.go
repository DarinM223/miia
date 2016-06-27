package graph

type GotoNode struct {
	id          int
	url         Node
	next        Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewGotoNode(id int, url Node) *GotoNode {
	return &GotoNode{
		id:          id,
		url:         url,
		inChan:      make(chan Msg, InChanSize),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *GotoNode) ID() int                       { return n.id }
func (n *GotoNode) Chan() chan Msg                { return n.inChan }
func (n *GotoNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *GotoNode) Destroy()                      {}
func (n *GotoNode) IsLoop() bool                  { return false }

func (n *GotoNode) Run() {
	for {
		msg := <-n.inChan
		if msg.Type == QuitMsg {
			break
		} else {
			var data Msg
			// TODO(DarinM223): send an HTTP request to get
			// and pass up the response.
			data.PassUp = true
			for _, parent := range n.parentChans {
				parent <- data
			}
			break
		}
	}
}