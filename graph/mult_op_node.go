package graph

import "github.com/DarinM223/http-scraper/tokens"

type MultOpNode struct {
	id          int
	operator    tokens.Token
	nodes       []Node
	inChan      chan Msg
	parentChans map[int]chan Msg
}

func NewMultOpNode(id int, operator tokens.Token, nodes []Node) *MultOpNode {
	inChan := make(chan Msg, InChanSize)
	for _, node := range nodes {
		node.ParentChans()[id] = inChan
	}

	return &MultOpNode{
		id:          id,
		operator:    operator,
		nodes:       nodes,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}
}

func (n *MultOpNode) ID() int                       { return n.id }
func (n *MultOpNode) Chan() chan Msg                { return n.inChan }
func (n *MultOpNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *MultOpNode) IsLoop() bool {
	isLoop := false
	for _, node := range n.nodes {
		if node.IsLoop() {
			isLoop = true
			break
		}
	}
	return isLoop
}

func (n *MultOpNode) Run() {
	var collectedData interface{} = nil
	var err error = nil

	passUpCount := 0
	for {
		msg := <-n.inChan
		if msg.Type == QuitMsg {
			for _, node := range n.nodes {
				node.Chan() <- msg
			}
			n.destroy()
			return
		} else if msg.PassUp {
			if collectedData == nil {
				collectedData = msg.Data
			} else {
				collectedData, err = applyMultOp(collectedData, msg.Data, n.operator)
				passUpCount++
				if err != nil || passUpCount >= len(n.nodes) {
					break
				}
			}
		}
	}

	var data Msg
	if err != nil {
		data = Msg{ErrMsg, true, err}
	} else {
		data = Msg{ValueMsg, true, collectedData}
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
	n.destroy()
}

func (n *MultOpNode) destroy() {
	for _, node := range n.nodes {
		delete(node.ParentChans(), n.id)
	}
}

func applyMultOp(collectedData interface{}, data interface{}, op tokens.Token) (interface{}, error) {
	return nil, nil
}
