package graph

type GotoNode struct {
	id          int
	next        Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewGotoNode(id int, next Node) *GotoNode {
	return &GotoNode{
		id:          id,
		next:        next,
		inChan:      make(chan Msg),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *GotoNode) ID() int                       { return n.id }
func (n *GotoNode) Chan() chan Msg                { return n.inChan }
func (n *GotoNode) ParentChans() map[int]chan Msg { return n.parentChans }

func (n *GotoNode) Run() {
	for {
		msg := <-n.inChan
		if msg.Type == QuitMsg {
			if n.next != nil {
				n.next.Chan() <- msg
			}
			n.Destroy()
			break
		} else if msg.PassUp {
			for _, parent := range n.parentChans {
				parent <- msg
			}
			n.Destroy()
			break
		} else {
			var data Msg
			// TODO(DarinM223): send an HTTP request to get
			// the page and then either send to next node
			// or pass up.
			if n.next != nil {
				n.next.Chan() <- data
			} else {
				data.PassUp = true
				for _, parent := range n.parentChans {
					parent <- data
				}
				n.Destroy()
				break
			}
		}
	}
}

func (n *GotoNode) AddChild(child Node) {
	if n.next != nil {
		n.RemoveChild(n.next)
	}
	n.next = child
	child.addParentChan(n.id, n.inChan)
}

func (n *GotoNode) RemoveChild(child Node) {
	if child.ID() == n.next.ID() {
		n.next.removeParentChan(n.id)
		n.next = nil
	}
}

func (n *GotoNode) Destroy() {
	if n.next != nil {
		n.RemoveChild(n.next)
		n.next = nil
	}
}

func (n *GotoNode) addParentChan(id int, parentChan chan Msg) { n.parentChans[id] = parentChan }
func (n *GotoNode) removeParentChan(id int)                   { delete(n.parentChans, id) }
