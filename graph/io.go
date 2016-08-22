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

func ReadNode(r io.Reader, g *Globals) (Node, error) {
	typeByte := make([]byte, 1)
	if _, err := r.Read(typeByte); err != nil {
		return nil, err
	}

	typeTag := NodeType(typeByte[0])
	switch typeTag {
	case BinOpType:
		return readBinOpNode(r, g)
	case CollectType:
		return readCollectNode(r, g)
	case ForType:
		return readForNode(r, g)
	case GotoType:
		return readGotoNode(r, g)
	case IfType:
		return readIfNode(r, g)
	case MultOpType:
		return readMultOpNode(r, g)
	case SelectorType:
		return readSelectorNode(r, g)
	case UnOpType:
		return readUnOpNode(r, g)
	case ValueType:
		return readValueNode(r, g)
	case VarType:
		return readVarNode(r, g)
	default:
		return nil, errors.New("Invalid node type")
	}
}

func readBinOpNode(r io.Reader, g *Globals) (*BinOpNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	a, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	b, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	return NewBinOpNode(g, id, tokens.Token(operator), a, b), nil
}

func readCollectNode(r io.Reader, g *Globals) (*CollectNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	node, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	return NewCollectNode(g, id, node), nil
}

func readForNode(r io.Reader, g *Globals) (*ForNode, error) {
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

	collection, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	body, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	forNode := NewForNode(g, id, name, collection, body)
	forNode.nodeType = nodeType.(forNodeType)
	forNode.setFanOut(fanout)
	return forNode, nil
}

func readGotoNode(r io.Reader, g *Globals) (*GotoNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	url, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	return NewGotoNode(g, id, url), nil
}

func readIfNode(r io.Reader, g *Globals) (*IfNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	pred, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	conseq, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	alt, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	return NewIfNode(g, id, pred, conseq, alt), nil
}

func readMultOpNode(r io.Reader, g *Globals) (*MultOpNode, error) {
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
	for i := 0; i < nodesLen; i++ {
		node, err := ReadNode(r, g)
		if err != nil {
			return nil, err
		}

		nodes[i] = node
	}

	return NewMultOpNode(g, id, tokens.Token(operator), nodes), nil
}

func readSelectorNode(r io.Reader, g *Globals) (*SelectorNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	gotoNode, err := ReadNode(r, g)
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

	return NewSelectorNode(g, id, gotoNode, selectors), nil
}

func readUnOpNode(r io.Reader, g *Globals) (*UnOpNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	node, err := ReadNode(r, g)
	if err != nil {
		return nil, err
	}

	return NewUnOpNode(g, id, tokens.Token(operator), node), nil
}

func readValueNode(r io.Reader, g *Globals) (*ValueNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	value, err := ReadInterface(r)
	if err != nil {
		return nil, err
	}

	return NewValueNode(g, id, value), nil
}

func readVarNode(r io.Reader, g *Globals) (*VarNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	name, err := ReadString(r)
	if err != nil {
		return nil, err
	}

	return NewVarNode(g, id, name), nil
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
