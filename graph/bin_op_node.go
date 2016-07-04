package graph

import (
	"errors"
	"github.com/DarinM223/http-scraper/tokens"
	"math"
	"reflect"
)

// BinOpNode listens to two nodes and applies
// an operator when receiving the values.
type BinOpNode struct {
	id           int
	operator     tokens.Token
	aChan, bChan chan Msg
	a, b         Node
	parentChans  map[int]chan Msg
}

func NewBinOpNode(globals *Globals, operator tokens.Token, a Node, b Node) *BinOpNode {
	id := globals.GenerateID()
	aChan := make(chan Msg, 1)
	bChan := make(chan Msg, 1)
	a.ParentChans()[id] = aChan
	b.ParentChans()[id] = bChan

	binOpNode := &BinOpNode{
		id:          id,
		operator:    operator,
		aChan:       aChan,
		bChan:       bChan,
		a:           a,
		b:           b,
		parentChans: make(map[int]chan Msg),
	}
	globals.RegisterNode(id, binOpNode)
	return binOpNode
}

func (n *BinOpNode) ID() int                       { return n.id }
func (n *BinOpNode) Chan() chan Msg                { return n.aChan }
func (n *BinOpNode) ParentChans() map[int]chan Msg { return n.parentChans }
func (n *BinOpNode) isLoop() bool                  { return n.a.isLoop() || n.b.isLoop() }

func (n *BinOpNode) Run() {
	val1 := <-n.aChan
	val2 := <-n.bChan

	var data Msg
	if val1.Type != ErrMsg && val2.Type != ErrMsg {
		result, err := applyBinOp(val1.Data, val2.Data, n.operator)
		if err != nil {
			data = Msg{ErrMsg, n.id, true, err}
		} else {
			data = Msg{ValueMsg, n.id, true, result}
		}
	} else {
		data = Msg{ErrMsg, n.id, true, errors.New("Error with BinOp values")}
	}

	for _, parent := range n.parentChans {
		parent <- data
	}
	n.destroy()
}

func (n *BinOpNode) Clone(globals *Globals) Node {
	clonedA, clonedB := n.a.Clone(globals), n.b.Clone(globals)
	retNode := NewBinOpNode(globals, n.operator, clonedA, clonedB)
	retNode.parentChans = n.parentChans
	return retNode
}

func (n *BinOpNode) destroy() {
	delete(n.a.ParentChans(), n.id)
	delete(n.b.ParentChans(), n.id)
}

func applyBinOp(a interface{}, b interface{}, op tokens.Token) (interface{}, error) {
	switch op {
	case tokens.RangeToken:
		firstVal, firstOk := a.(int)
		secondVal, secondOk := b.(int)

		if firstOk && secondOk {
			rangeInts := make([]int, int(math.Abs(float64(secondVal-firstVal))))
			for i := 0; i < len(rangeInts); i++ {
				if secondVal > firstVal {
					rangeInts[i] = firstVal + i
				} else {
					rangeInts[i] = firstVal - i
				}
			}
			return rangeInts, nil
		}
		return nil, errors.New("Invalid types for BinOp RangeToken")
	case tokens.EqualsToken:
		if reflect.TypeOf(a) != reflect.TypeOf(b) {
			return nil, errors.New("Invalid types for BinOp EqualsToken")
		}
		return reflect.DeepEqual(a, b), nil
	case tokens.OrToken:
		firstVal, firstOk := a.(bool)
		secondVal, secondOk := b.(bool)

		if firstOk && secondOk {
			return firstVal || secondVal, nil
		}
		return nil, errors.New("Invalid types for BinOp OrToken")
	case tokens.AndToken:
		firstVal, firstOk := a.(bool)
		secondVal, secondOk := b.(bool)

		if firstOk && secondOk {
			return firstVal && secondVal, nil
		}
		return nil, errors.New("Invalid types for BinOp AndToken")
	default:
		return nil, errors.New("Invalid BinOp operator")
	}
}
