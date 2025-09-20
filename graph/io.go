package graph

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/DarinM223/miia/tokens"
	"golang.org/x/time/rate"
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
	if n, err := r.Read(buf); err != nil {
		return -1, fmt.Errorf("error reading to 32-bit integer buffer (%d bytes read): %w", n, err)
	}

	return int(buf[0]) + (int(buf[1]) << 8) + (int(buf[2]) << 16) + (int(buf[3]) << 24), nil
}

func ReadInt64(r io.Reader) (int64, error) {
	buf := make([]byte, 8)
	if n, err := r.Read(buf); err != nil {
		return -1, fmt.Errorf("error reading to 64-bit integer buffer (%d bytes read): %w", n, err)
	}

	return int64(buf[0]) + (int64(buf[1]) << 8) + (int64(buf[2]) << 16) + (int64(buf[3]) << 24) +
		(int64(buf[4]) << 32) + (int64(buf[5]) << 40) + (int64(buf[6]) << 48) + (int64(buf[7]) << 56), nil
}

func ReadString(r io.Reader) (string, error) {
	len, err := ReadInt(r)
	if err != nil {
		return "", fmt.Errorf("error reading string length: %w", err)
	}

	buf := make([]byte, len)
	if n, err := r.Read(buf); err != nil {
		return "", fmt.Errorf("error reading to string buffer (%d bytes read): %w", n, err)
	}

	return string(buf[:]), nil
}

func ReadValue(r io.Reader) (result any, err error) {
	dataType, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading value tag: %w", err)
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

	if n, err := w.Write([]byte{b0, b1, b2, b3}); err != nil {
		return fmt.Errorf("error writing 32-bit integer (wrote %d bytes): %w", n, err)
	}
	return nil
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

	if n, err := w.Write([]byte{b0, b1, b2, b3, b4, b5, b6, b7}); err != nil {
		return fmt.Errorf("error writing 64-bit integer (wrote %d bytes): %w", n, err)
	}
	return nil
}

func WriteString(w io.Writer, s string) error {
	if err := WriteInt(w, len(s)); err != nil {
		return fmt.Errorf("error writing string length: %w", err)
	}

	bytes := []byte(s)
	if n, err := w.Write(bytes); err != nil {
		return fmt.Errorf("error writing string (wrote %d bytes): %w", n, err)
	}
	return nil
}

func WriteValue(w io.Writer, i any) error {
	errInvalidType := errors.New("invalid data type to encode")
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
			return errInvalidType
		}
	}

	if err := WriteInt(w, int(dataType)); err != nil {
		return fmt.Errorf("error writing value tag: %w", err)
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
		return errInvalidType
	}
}

/*
 * Implementations for reading nodes from files.
 */

func ReadGlobals(r io.Reader) (*Globals, error) {
	currID, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading globals current id: %w", err)
	}

	resultNodeID, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading globals result node id: %w", err)
	}

	rateLimitersLen, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading globals rate limiters length: %w", err)
	}

	globals := &Globals{
		currID:          currID,
		resultID:        resultNodeID,
		mutex:           &sync.Mutex{},
		nodeMap:         make(map[int]Node),
		rateLimiterData: make(map[string]RateLimiterData),
		rateLimiters:    make(map[string]*rate.Limiter),
	}

	for i := 0; i < rateLimitersLen; i++ {
		domain, err := ReadString(r)
		if err != nil {
			return nil, fmt.Errorf("error reading rate limiter domain for index %d: %w", i, err)
		}

		limit, err := ReadInt(r)
		if err != nil {
			return nil, fmt.Errorf("error reading rate limiter limit for index %d: %w", i, err)
		}

		duration, err := ReadInt64(r)
		if err != nil {
			return nil, fmt.Errorf("error reading rate limiter duration for index %d: %w", i, err)
		}

		globals.SetRateLimit(domain, limit, time.Duration(duration))
	}

	// Read result node.
	if _, err := ReadNode(r, globals); err != nil {
		return nil, fmt.Errorf("error reading globals result node: %w", err)
	}
	return globals, nil
}

func ReadNode(r io.Reader, g *Globals) (Node, error) {
	typeByte := make([]byte, 1)
	if _, err := r.Read(typeByte); err != nil {
		return nil, fmt.Errorf("error reading node tag: %w", err)
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
		return nil, fmt.Errorf("error reading binop id: %w", err)
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading binop operator: %w", err)
	}

	a, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading binop lhs: %w", err)
	}

	b, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading binop rhs: %w", err)
	}

	return NewBinOpNode(g, id, tokens.Token(operator), a, b), nil
}

func readCollectNode(r io.Reader, g *Globals) (*CollectNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading collect id: %w", err)
	}

	node, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading collect node: %w", err)
	}

	return NewCollectNode(g, id, node), nil
}

func readForNode(r io.Reader, g *Globals) (*ForNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading for node id: %w", err)
	}

	fanout, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading for node fanout: %w", err)
	}

	name, err := ReadString(r)
	if err != nil {
		return nil, fmt.Errorf("error reading for node name: %w", err)
	}

	collection, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading for node collection: %w", err)
	}

	body, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading for node body: %w", err)
	}

	forNode := NewForNode(g, id, name, collection, body)
	forNode.setFanOut(fanout)
	return forNode, nil
}

func readGotoNode(r io.Reader, g *Globals) (*GotoNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading goto id: %w", err)
	}

	url, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading goto url: %w", err)
	}

	return NewGotoNode(g, id, url), nil
}

func readIfNode(r io.Reader, g *Globals) (*IfNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading if node id: %w", err)
	}

	pred, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading if node predicate: %w", err)
	}

	conseq, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading if node consequence: %w", err)
	}

	alt, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading if node alternative: %w", err)
	}

	return NewIfNode(g, id, pred, conseq, alt), nil
}

func readMultOpNode(r io.Reader, g *Globals) (*MultOpNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading multop id: %w", err)
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading multop operator: %w", err)
	}

	nodesLen, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading multop nodes length: %w", err)
	}

	nodes := make([]Node, nodesLen)
	for i := 0; i < nodesLen; i++ {
		node, err := ReadNode(r, g)
		if err != nil {
			return nil, fmt.Errorf("error reading multop node at index %d: %w", i, err)
		}

		nodes[i] = node
	}

	return NewMultOpNode(g, id, tokens.Token(operator), nodes), nil
}

func readSelectorNode(r io.Reader, g *Globals) (*SelectorNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading selector node id: %w", err)
	}

	gotoNode, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading selector node goto: %w", err)
	}

	selectorsLen, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading selector node length: %w", err)
	}

	selectors := make([]Selector, selectorsLen)
	for i := 0; i < selectorsLen; i++ {
		name, err := ReadString(r)
		if err != nil {
			return nil, fmt.Errorf("error reading selector name at index %d: %w", i, err)
		}

		selector, err := ReadString(r)
		if err != nil {
			return nil, fmt.Errorf("error reading selector for %s at index %d: %w", name, i, err)
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
		return nil, fmt.Errorf("error reading unop id: %w", err)
	}

	operator, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading unop operator: %w", err)
	}

	node, err := ReadNode(r, g)
	if err != nil {
		return nil, fmt.Errorf("error reading unop node: %w", err)
	}

	return NewUnOpNode(g, id, tokens.Token(operator), node), nil
}

func readValueNode(r io.Reader, g *Globals) (*ValueNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading value node id: %w", err)
	}

	value, err := ReadValue(r)
	if err != nil {
		return nil, fmt.Errorf("error reading value node value: %w", err)
	}

	return NewValueNode(g, id, value), nil
}

func readVarNode(r io.Reader, g *Globals) (*VarNode, error) {
	id, err := ReadInt(r)
	if err != nil {
		return nil, fmt.Errorf("error reading var node id: %w", err)
	}

	name, err := ReadString(r)
	if err != nil {
		return nil, fmt.Errorf("error reading var node name: %w", err)
	}

	return NewVarNode(g, id, name), nil
}

/*
 * Implementations for writing nodes to files.
 */

func WriteGlobals(w io.Writer, g *Globals) error {
	if err := WriteInt(w, g.currID); err != nil {
		return fmt.Errorf("error writing globals current id: %w", err)
	}
	if err := WriteInt(w, g.resultID); err != nil {
		return fmt.Errorf("error writing globals result id: %w", err)
	}
	if err := WriteInt(w, len(g.rateLimiterData)); err != nil {
		return fmt.Errorf("error writing globals rate limiter length: %w", err)
	}

	for domain, rateLimiter := range g.rateLimiterData {
		if err := WriteString(w, domain); err != nil {
			return fmt.Errorf("error writing domain %s: %w", domain, err)
		}
		if err := WriteInt(w, rateLimiter.limit); err != nil {
			return fmt.Errorf("error writing rate limiter limit for domain %s: %w", domain, err)
		}
		if err := WriteInt64(w, int64(rateLimiter.interval)); err != nil {
			return fmt.Errorf("error writing rate limiter interval for domain %s: %w", domain, err)
		}
	}

	if err := g.ResultNode().Write(w); err != nil {
		return fmt.Errorf("error writing globals result node: %w", err)
	}
	return nil
}

func (n *BinOpNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(BinOpType)}); err != nil {
		return fmt.Errorf("error writing binop tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing binop id: %w", err)
	}
	if err := WriteInt(w, int(n.operator)); err != nil {
		return fmt.Errorf("error writing binop operator: %w", err)
	}
	if err := n.a.Write(w); err != nil {
		return fmt.Errorf("error writing binop lhs: %w", err)
	}
	if err := n.b.Write(w); err != nil {
		return fmt.Errorf("error writing binop rhs: %w", err)
	}
	return nil
}

func (n *CollectNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(CollectType)}); err != nil {
		return fmt.Errorf("error writing collect tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing collect id: %w", err)
	}
	if err := n.node.Write(w); err != nil {
		return fmt.Errorf("error writing collect node: %w", err)
	}
	return nil
}

func (n *ForNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(ForType)}); err != nil {
		return fmt.Errorf("error writing for node tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing for node id: %w", err)
	}
	if err := WriteInt(w, n.fanout); err != nil {
		return fmt.Errorf("error writing for node fanout: %w", err)
	}
	if err := WriteString(w, n.name); err != nil {
		return fmt.Errorf("error writing for node name: %w", err)
	}
	if err := n.collection.Write(w); err != nil {
		return fmt.Errorf("error writing for node collection: %w", err)
	}
	if err := n.body.Write(w); err != nil {
		return fmt.Errorf("error writing for node body: %w", err)
	}
	return nil
}

func (n *GotoNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(GotoType)}); err != nil {
		return fmt.Errorf("error writing goto tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing goto id: %w", err)
	}
	if err := n.url.Write(w); err != nil {
		return fmt.Errorf("error writing goto url: %w", err)
	}
	return nil
}

func (n *IfNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(IfType)}); err != nil {
		return fmt.Errorf("error writing if node tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing if node id: %w", err)
	}
	if err := n.pred.Write(w); err != nil {
		return fmt.Errorf("error writing if node predicate: %w", err)
	}
	if err := n.conseq.Write(w); err != nil {
		return fmt.Errorf("error writing if node consequence: %w", err)
	}
	if err := n.alt.Write(w); err != nil {
		return fmt.Errorf("error writing if node alternative: %w", err)
	}
	return nil
}

func (n *MultOpNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(MultOpType)}); err != nil {
		return fmt.Errorf("error writing multop tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing multop id: %w", err)
	}
	if err := WriteInt(w, int(n.operator)); err != nil {
		return fmt.Errorf("error writing multop operator: %w", err)
	}
	if err := WriteInt(w, len(n.nodes)); err != nil {
		return fmt.Errorf("error writing multop nodes length: %w", err)
	}
	for i, node := range n.nodes {
		if err := node.Write(w); err != nil {
			return fmt.Errorf("error writing multop node at index %d: %w", i, err)
		}
	}
	return nil
}

func (n *SelectorNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(SelectorType)}); err != nil {
		return fmt.Errorf("error writing selector node tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing selector node id: %w", err)
	}
	if err := n.gotoNode.Write(w); err != nil {
		return fmt.Errorf("error writing selector node goto: %w", err)
	}
	if err := WriteInt(w, len(n.selectors)); err != nil {
		return fmt.Errorf("error writing selector node length: %w", err)
	}
	for i, selector := range n.selectors {
		if err := WriteString(w, selector.Name); err != nil {
			return fmt.Errorf("error writing selector name at index %d: %w", i, err)
		}
		if err := WriteString(w, selector.Selector); err != nil {
			return fmt.Errorf("error writing selector for %s at index %d: %w", selector.Name, i, err)
		}
	}
	return nil
}

func (n *UnOpNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(UnOpType)}); err != nil {
		return fmt.Errorf("error writing unop tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing unop id: %w", err)
	}
	if err := WriteInt(w, int(n.operator)); err != nil {
		return fmt.Errorf("error writing unop operator: %w", err)
	}
	if err := n.node.Write(w); err != nil {
		return fmt.Errorf("error writing unop node: %w", err)
	}
	return nil
}

func (n *ValueNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(ValueType)}); err != nil {
		return fmt.Errorf("error writing value node tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing value node id: %w", err)
	}
	if err := WriteValue(w, n.value); err != nil {
		return fmt.Errorf("error writing value node value: %w", err)
	}
	return nil
}

func (n *VarNode) Write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(VarType)}); err != nil {
		return fmt.Errorf("error writing var node tag: %w", err)
	}

	if err := WriteInt(w, n.id); err != nil {
		return fmt.Errorf("error writing var node id: %w", err)
	}
	if err := WriteString(w, n.name); err != nil {
		return fmt.Errorf("error writing var node name: %w", err)
	}
	return nil
}
