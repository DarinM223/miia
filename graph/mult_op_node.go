package graph

import (
	"errors"
	"github.com/DarinM223/http-scraper/tokens"
)

// MultOpNode listens to multiple values and
// applies an operator to the values when all
// of them are received.
type MultOpNode struct {
	id          int
	operator    tokens.Token
	nodes       []Node
	inChan      chan Msg
	parentChans map[int]chan Msg
	// Stores the results of the nodes.
	// Index of a result corresponds to the index
	// of the node in nodes that created the result.
	results []interface{}
	// Map from the ID of the node to the
	// index of the node from nodes.
	idMap map[int]int
}

func NewMultOpNode(id int, operator tokens.Token, nodes []Node) *MultOpNode {
	inChan := make(chan Msg, len(nodes))
	idMap := make(map[int]int, len(nodes))
	for i, node := range nodes {
		node.ParentChans()[id] = inChan
		idMap[node.ID()] = i
	}

	return &MultOpNode{
		id:          id,
		operator:    operator,
		nodes:       nodes,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
		results:     make([]interface{}, len(nodes)),
		idMap:       idMap,
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
			// Store the result and break when all the nodes have finished.
			nodeIdx := n.idMap[msg.ID]
			n.results[nodeIdx] = msg.Data
			passUpCount++
			if passUpCount >= len(n.nodes) {
				break
			}
		}
	}

	var msg Msg
	data, err := applyMultOp(n.results, n.operator)
	if err != nil {
		msg = Msg{ErrMsg, n.id, true, err}
	} else {
		msg = Msg{ValueMsg, n.id, true, data}
	}

	for _, parent := range n.parentChans {
		parent <- msg
	}
	n.destroy()
}

func (n *MultOpNode) destroy() {
	for _, node := range n.nodes {
		delete(node.ParentChans(), n.id)
	}
}

func applyMultOp(data []interface{}, op tokens.Token) (interface{}, error) {
	if len(data) <= 0 {
		return nil, errors.New("Need to apply MultOp to at least one element")
	}
	switch op {
	case tokens.AddToken:
		switch result := data[0].(type) {
		case string:
			for i := 1; i < len(data); i++ {
				elem, ok := data[i].(string)
				if !ok {
					return nil, errors.New("Invalid types for MultOp AddToken")
				}
				result += elem
			}
			return result, nil
		case int:
			for i := 1; i < len(data); i++ {
				elem, ok := data[i].(int)
				if !ok {
					return nil, errors.New("Invalid types for MultOp AddToken")
				}
				result += elem
			}
			return result, nil
		default:
			return nil, errors.New("Invalid types for MultOp AddToken")
		}
	case tokens.SubToken:
		result, ok := data[0].(int)
		if !ok {
			return nil, errors.New("Invalid types for MultOp SubToken")
		}

		for i := 1; i < len(data); i++ {
			elem, ok := data[i].(int)
			if !ok {
				return nil, errors.New("Invalid types for MultOp SubToken")
			}
			result -= elem
		}
		return result, nil
	case tokens.MulToken:
		result := 1
		for i := 0; i < len(data); i++ {
			elem, ok := data[i].(int)
			if !ok {
				return nil, errors.New("Invalid types for MultOp MulToken")
			}
			result *= elem
		}
		return result, nil
	case tokens.DivToken:
		result, ok := data[0].(int)
		if !ok {
			return nil, errors.New("Invalid types for MultOp Divtoken")
		}

		for i := 1; i < len(data); i++ {
			elem, ok := data[i].(int)
			if !ok {
				return nil, errors.New("Invalid types for MultOp Divtoken")
			}
			result /= elem
		}
		return result, nil
	default:
		return nil, errors.New("Invalid MultOp operator")
	}
}
