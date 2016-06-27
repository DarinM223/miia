package graph

type SelectorType int

const (
	SelectorClass SelectorType = iota // A CSS class to retrieve.
	SelectorID                        // A CSS id to retrieve.
)

// SelectorNode is a node that receives HTTP Responses and parses out CSS selectors
// and passes the result back to the parents.
type SelectorNode struct {
	id          int
	selType     SelectorType
	selector    string
	gotoNode    Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewSelectorNode(id int, gotoNode Node, selType SelectorType, selector string) *SelectorNode {
	return &SelectorNode{
		id:          id,
		selType:     selType,
		selector:    selector,
		gotoNode:    gotoNode,
		inChan:      make(chan Msg, InChanSize),
		parentChans: make(map[int]chan Msg),
	}
}

func (n *SelectorNode) ID() int                       { return n.id }
func (n *SelectorNode) Chan() chan Msg                { return n.inChan }
func (n *SelectorNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *SelectorNode) IsLoop() bool                  { return false }

func (n *SelectorNode) Run() {
	// TODO(DarinM223): listen for HTTP responses and then parse them
}

func (n *SelectorNode) Destroy() {
	if n.gotoNode != nil {
		n.gotoNode.removeParentChan(n.id)
		n.gotoNode = nil
	}
}

func (n *SelectorNode) addParentChan(id int, parentChan chan Msg) { n.parentChans[id] = parentChan }
func (n *SelectorNode) removeParentChan(id int)                   { delete(n.parentChans, id) }
