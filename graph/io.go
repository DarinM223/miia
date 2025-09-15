package graph

import (
	"errors"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/DarinM223/miia/tokens"
	"github.com/beefsack/go-rate"
)

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func Itob(i int) bool {
	return i != 0
}

/*
 * Helper functions for reading and writing to files.
 */

type DataType int

const (
	IntType DataType = iota
	StringType
	BoolType
	NilType
)

func ReadInt(r io.Reader) (int, error) {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return -1, err
	}

	return int(buf[0]) + (int(buf[1]) << 8) + (int(buf[2]) << 16) + (int(buf[3]) << 24), nil
}

func ReadInt64(r io.Reader) (int64, error) {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return -1, err
	}

	return int64(buf[0]) + (int64(buf[1]) << 8) + (int64(buf[2]) << 16) + (int64(buf[3]) << 24) +
		(int64(buf[4]) << 32) + (int64(buf[5]) << 40) + (int64(buf[6]) << 48) + (int64(buf[7]) << 56), nil
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

func ReadValue(r io.Reader) (result any, err error) {
	dataType, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	switch DataType(dataType) {
	case IntType:
		result, err = ReadInt(r)
	case BoolType:
		boolNum, e := ReadInt(r)
		result, err = Itob(boolNum), e
	case StringType:
		result, err = ReadString(r)
	case NilType:
		result, err = nil, nil
	default:
		err = errors.New("invalid data type for decoding")
	}
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

func WriteInt64(w io.Writer, i int64) error {
	b7 := byte((i >> 56) & (0xFF))
	b6 := byte((i >> 48) & (0xFF))
	b5 := byte((i >> 40) & (0xFF))
	b4 := byte((i >> 32) & (0xFF))
	b3 := byte((i >> 24) & (0xFF))
	b2 := byte((i >> 16) & (0xFF))
	b1 := byte((i >> 8) & (0xFF))
	b0 := byte(i & 0xFF)

	_, err := w.Write([]byte{b0, b1, b2, b3, b4, b5, b6, b7})
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

func WriteValue(w io.Writer, i any) error {
	err := errors.New("invalid data type to encode")
	var dataType DataType
	if i == nil {
		dataType = NilType
	} else {
		switch reflect.TypeOf(i).Kind() {
		case reflect.Int:
			dataType = IntType
		case reflect.Bool:
			dataType = BoolType
		case reflect.String:
			dataType = StringType
		default:
			return err
		}
	}

	if err := WriteInt(w, int(dataType)); err != nil {
		return err
	}

	switch dataType {
	case IntType:
		return WriteInt(w, i.(int))
	case BoolType:
		return WriteInt(w, Btoi(i.(bool)))
	case StringType:
		return WriteString(w, i.(string))
	case NilType:
		return nil
	default:
		return err
	}
}

/*
 * Implementations for reading nodes from files.
 */

func ReadGlobals(r io.Reader) (*Globals, error) {
	currID, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	resultNodeID, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	rateLimitersLen, err := ReadInt(r)
	if err != nil {
		return nil, err
	}

	globals := &Globals{
		currID:          currID,
		resultID:        resultNodeID,
		mutex:           &sync.Mutex{},
		nodeMap:         make(map[int]Node),
		rateLimiterData: make(map[string]RateLimiterData),
		rateLimiters:    make(map[string]*rate.RateLimiter),
	}

	for i := 0; i < rateLimitersLen; i++ {
		domain, err := ReadString(r)
		if err != nil {
			return nil, err
		}

		limit, err := ReadInt(r)
		if err != nil {
			return nil, err
		}

		duration, err := ReadInt64(r)
		if err != nil {
			return nil, err
		}

		globals.SetRateLimit(domain, limit, time.Duration(duration))
	}

	// Read result node.
	if _, err := ReadNode(r, globals); err != nil {
		return nil, err
	}
	return globals, nil
}

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
		return nil, errors.New("invalid node type")
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

	value, err := ReadValue(r)
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

func WriteGlobals(w io.Writer, g *Globals) error {
	if err := WriteInt(w, g.currID); err != nil {
		return err
	}
	if err := WriteInt(w, g.resultID); err != nil {
		return err
	}
	if err := WriteInt(w, len(g.rateLimiterData)); err != nil {
		return err
	}

	for domain, rateLimiter := range g.rateLimiterData {
		if err := WriteString(w, domain); err != nil {
			return err
		}
		if err := WriteInt(w, rateLimiter.limit); err != nil {
			return err
		}
		if err := WriteInt64(w, int64(rateLimiter.interval)); err != nil {
			return err
		}
	}

	if err := g.ResultNode().Write(w); err != nil {
		return err
	}
	return nil
}

func (n *BinOpNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(BinOpType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := WriteInt(w, int(n.operator)); err != nil {
		return err
	}
	if err := n.a.Write(w); err != nil {
		return err
	}
	if err := n.b.Write(w); err != nil {
		return err
	}
	return nil
}

func (n *CollectNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(CollectType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := n.node.Write(w); err != nil {
		return err
	}
	return nil
}

func (n *ForNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(ForType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := WriteInt(w, n.fanout); err != nil {
		return err
	}
	if err := WriteString(w, n.name); err != nil {
		return err
	}
	if err := n.collection.Write(w); err != nil {
		return err
	}
	if err := n.body.Write(w); err != nil {
		return err
	}
	return nil
}

func (n *GotoNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(GotoType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := n.url.Write(w); err != nil {
		return err
	}
	return nil
}

func (n *IfNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(IfType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := n.pred.Write(w); err != nil {
		return err
	}
	if err := n.conseq.Write(w); err != nil {
		return err
	}
	if err := n.alt.Write(w); err != nil {
		return err
	}
	return nil
}

func (n *MultOpNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(MultOpType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := WriteInt(w, int(n.operator)); err != nil {
		return err
	}
	if err := WriteInt(w, len(n.nodes)); err != nil {
		return err
	}
	for _, node := range n.nodes {
		if err := node.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (n *SelectorNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(SelectorType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := n.gotoNode.Write(w); err != nil {
		return err
	}
	if err := WriteInt(w, len(n.selectors)); err != nil {
		return err
	}
	for _, selector := range n.selectors {
		if err := WriteString(w, selector.Name); err != nil {
			return err
		}
		if err := WriteString(w, selector.Selector); err != nil {
			return err
		}
	}
	return nil
}

func (n *UnOpNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(UnOpType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := WriteInt(w, int(n.operator)); err != nil {
		return err
	}
	if err := n.node.Write(w); err != nil {
		return err
	}
	return nil
}

func (n *ValueNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(ValueType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := WriteValue(w, n.value); err != nil {
		return err
	}
	return nil
}

func (n *VarNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(VarType)}); err != nil {
		return err
	}

	if err := WriteInt(w, n.id); err != nil {
		return err
	}
	if err := WriteString(w, n.name); err != nil {
		return err
	}
	return nil
}
