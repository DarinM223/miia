package graph

import (
	"encoding/gob"
	"errors"
	"github.com/DarinM223/miia/tokens"
	"io"
)

/*
 * Helper functions for reading and writing to files.
 */

func ReadInt(r io.Reader) (int, error) {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return -1, err
	}

	return int(buf[0]) + (int(buf[1]) << 8) + (int(buf[2]) << 16) + (int(buf[3]) << 24), nil
}

func ReadString(r io.Reader) (string, error) {
	len, err := ReadInt(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, len)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}

	return string(buf[:]), nil
}

func ReadInterface(r io.Reader) (result interface{}, err error) {
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&result)
	return
}

func WriteInt(w io.Writer, i int) error {
	b3 := byte((i >> 24) & (0xFF))
	b2 := byte((i >> 16) & (0xFF))
	b1 := byte((i >> 8) & (0xFF))
	b0 := byte(i & 0xFF)

	_, err := w.Write([]byte{b0, b1, b2, b3})
	return err
}

func WriteString(w io.Writer, s string) error {
	if err := WriteInt(w, len(s)); err != nil {
		return err
	}

	bytes := []byte(s)
	_, err := w.Write(bytes)
	return err
}

func WriteInterface(w io.Writer, i interface{}) error {
	encoder := gob.NewEncoder(w)
	return encoder.Encode(&i)
}

/*
 * Implementations for reading nodes from files.
 */

func ReadNode(r io.Reader) (Node, error) {
	typeByte := make([]byte, 1)
	if _, err := r.Read(typeByte); err != nil {
		return nil, err
	}

	typeTag := NodeType(typeByte[0])
	switch typeTag {
	case BinOpType:
		return readBinOpNode(r)
	case CollectType:
		return readCollectNode(r)
	case ForType:
		return readForNode(r)
	case GotoType:
		return readGotoNode(r)
	case IfType:
		return readIfNode(r)
	case MultOpType:
		return readMultOpNode(r)
	case SelectorType:
		return readSelectorNode(r)
	case UnOpType:
		return readUnOpNode(r)
	case ValueType:
		return readValueNode(r)
	case VarType:
		return readVarNode(r)
	default:
		return nil, errors.New("Invalid node type")
	}
}

func readBinOpNode(r io.Reader) (*BinOpNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	a, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	b, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	aChan := make(chan Msg, 1)
	bChan := make(chan Msg, 1)
	a.ParentChans()[id] = aChan
	b.ParentChans()[id] = bChan

	return &BinOpNode{
		id:          id,
		operator:    tokens.Token(operator),
		aChan:       aChan,
		bChan:       bChan,
		a:           a,
		b:           b,
		parentChans: make(map[int]chan Msg),
	}, nil
}

func readCollectNode(r io.Reader) (*CollectNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	node, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	inChan := make(chan Msg, InChanSize)
	node.ParentChans()[id] = inChan

	return &CollectNode{
		id:          id,
		node:        node,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
		results:     nil,
	}, nil
}

func readForNode(r io.Reader) (*ForNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	fanout, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	nodeType, err := ReadInterface(r)
	if err != nil {
		return nil, err
	}

	name, err := ReadString(r)
	if err != nil {
		return nil, err
	}

	collection, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	body, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	collectionChan := make(chan Msg, InChanSize)
	collection.ParentChans()[id] = collectionChan

	return &ForNode{
		id:             id,
		fanout:         fanout,
		nodeType:       nodeType.(forNodeType),
		subnodes:       make(map[string]Node),
		name:           name,
		collection:     collection,
		body:           body,
		inChan:         nil,
		collectionChan: collectionChan,
		parentChans:    make(map[int]chan Msg),
		nodeToIdx:      make(map[int]StreamIndex),
		isLoop:         ContainsLoopNode(body),
	}, nil
}

func readGotoNode(r io.Reader) (*GotoNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	url, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	inChan := make(chan Msg, InChanSize)
	url.ParentChans()[id] = inChan

	return &GotoNode{
		id:          id,
		url:         url,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}, nil
}

func readIfNode(r io.Reader) (*IfNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	pred, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	conseq, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	alt, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	inChan := make(chan Msg, InChanSize)
	conseqChan := make(chan Msg, InChanSize)
	altChan := make(chan Msg, InChanSize)

	pred.ParentChans()[id] = inChan
	conseq.ParentChans()[id] = conseqChan
	alt.ParentChans()[id] = altChan

	return &IfNode{
		id:          id,
		pred:        pred,
		conseq:      conseq,
		alt:         alt,
		inChan:      inChan,
		conseqChan:  conseqChan,
		altChan:     altChan,
		parentChans: make(map[int]chan Msg),
	}, nil
}

func readMultOpNode(r io.Reader) (*MultOpNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	nodesLen, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, nodesLen)
	inChan := make(chan Msg, len(nodes))
	idMap := make(map[int]int, len(nodes))

	for i := 0; i < nodesLen; i++ {
		node, err := ReadNode(r)
		if err != nil {
			return nil, err
		}

		node.ParentChans()[id] = inChan
		idMap[node.ID()] = i

		nodes[i] = node
	}

	return &MultOpNode{
		id:          id,
		operator:    tokens.Token(operator),
		nodes:       nodes,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
		results:     make([]interface{}, len(nodes)),
		idMap:       idMap,
	}, nil
}

func readSelectorNode(r io.Reader) (*SelectorNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	gotoNode, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	selectorsLen, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	selectors := make([]Selector, selectorsLen)
	for i := 0; i < selectorsLen; i++ {
		name, err := ReadString(r)
		if err != nil {
			return nil, err
		}

		selector, err := ReadString(r)
		if err != nil {
			return nil, err
		}

		selectors[i] = Selector{
			Name:     name,
			Selector: selector,
		}
	}

	inChan := make(chan Msg, InChanSize)
	gotoNode.ParentChans()[id] = inChan

	return &SelectorNode{
		id:          id,
		selectors:   selectors,
		gotoNode:    gotoNode,
		inChan:      inChan,
		parentChans: make(map[int]chan Msg),
	}, nil
}

func readUnOpNode(r io.Reader) (*UnOpNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	node, err := ReadNode(r)
	if err != nil {
		return nil, err
	}

	inChan := make(chan Msg, 1)
	node.ParentChans()[id] = inChan

	return &UnOpNode{
		id:          id,
		operator:    tokens.Token(operator),
		inChan:      inChan,
		node:        node,
		parentChans: make(map[int]chan Msg),
	}, nil
}

func readValueNode(r io.Reader) (*ValueNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	value, err := ReadInterface(r)
	if err != nil {
		return nil, err
	}

	return &ValueNode{
		id:          id,
		value:       value,
		inChan:      make(chan Msg, InChanSize),
		parentChans: make(map[int]chan Msg),
	}, nil
}

func readVarNode(r io.Reader) (*VarNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	name, err := ReadString(r)
	if err != nil {
		return nil, err
	}

	return &VarNode{
		id:          id,
		name:        name,
		msg:         nil,
		inChan:      make(chan Msg, 1),
		parentChans: make(map[int]chan Msg),
	}, nil
}

/*
 * Implementations for writing nodes to files.
 */

func (n *BinOpNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *CollectNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *ForNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *GotoNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *IfNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *MultOpNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *SelectorNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *UnOpNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *ValueNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *VarNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}
