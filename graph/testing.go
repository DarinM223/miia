package graph

import (
	"github.com/DarinM223/http-scraper/tokens"
	"reflect"
)

// Testing contains utility methods for generating nodes
// inside test functions without needing the global node map.
type Testing struct{}

var testUtils Testing = Testing{}

func (t *Testing) CompareTestNodeToRealNode(testNode Node, realNode Node) bool {
	if reflect.TypeOf(testNode) != reflect.TypeOf(realNode) {
		return false
	}

	switch n := testNode.(type) {
	case *BinOpNode:
		compare := realNode.(*BinOpNode)
		compareA := t.CompareTestNodeToRealNode(n.a, compare.a)
		compareB := t.CompareTestNodeToRealNode(n.b, compare.b)
		return n.operator == compare.operator && compareA && compareB
	case *ForNode:
		compare := realNode.(*ForNode)
		compareColl := t.CompareTestNodeToRealNode(n.collection, compare.collection)
		compareBody := t.CompareTestNodeToRealNode(n.body, compare.body)
		return n.name == compare.name && compareColl && compareBody
	case *GotoNode:
		return t.CompareTestNodeToRealNode(n.url, realNode.(*GotoNode).url)
	case *IfNode:
		compare := realNode.(*IfNode)
		comparePred := t.CompareTestNodeToRealNode(n.pred, compare.pred)
		compareConseq := t.CompareTestNodeToRealNode(n.conseq, compare.conseq)
		compareAlt := t.CompareTestNodeToRealNode(n.alt, compare.alt)
		return comparePred && compareConseq && compareAlt
	case *MultOpNode:
		compare := realNode.(*MultOpNode)
		if len(n.nodes) != len(compare.nodes) {
			return false
		}

		compareNodes := true
		for i := 0; i < len(n.nodes); i++ {
			if !t.CompareTestNodeToRealNode(n.nodes[i], compare.nodes[i]) {
				compareNodes = false
				break
			}
		}
		return n.operator == compare.operator && compareNodes
	case *SelectorNode:
		compare := realNode.(*SelectorNode)
		compareGoto := t.CompareTestNodeToRealNode(n.gotoNode, compare.gotoNode)
		return compareGoto && reflect.DeepEqual(n.selectors, compare.selectors)
	case *UnOpNode:
		compare := realNode.(*UnOpNode)
		return n.operator == compare.operator && t.CompareTestNodeToRealNode(n.node, compare.node)
	case *ValueNode:
		return reflect.DeepEqual(n.value, realNode.(*ValueNode).value)
	case *VarNode:
		return n.name == realNode.(*VarNode).name
	default:
		panic("Invalid node type for CompareTestNodeToRealNode()")
	}
}

func (t *Testing) GenerateTestNode(g *Globals, node Node) Node {
	if _, ok := g.nodeMap[node.ID()]; ok && node.ID() != -1 {
		return g.nodeMap[node.ID()]
	}

	switch n := node.(type) {
	case *BinOpNode:
		a, b := t.GenerateTestNode(g, n.a), t.GenerateTestNode(g, n.b)
		return NewBinOpNode(g, n.operator, a, b)
	case *ForNode:
		collection, body := t.GenerateTestNode(g, n.collection), t.GenerateTestNode(g, n.body)
		return NewForNode(g, n.name, collection, body)
	case *GotoNode:
		url := t.GenerateTestNode(g, n.url)
		return NewGotoNode(g, url)
	case *IfNode:
		pred, conseq, alt := t.GenerateTestNode(g, n.pred), t.GenerateTestNode(g, n.conseq), t.GenerateTestNode(g, n.alt)
		return NewIfNode(g, pred, conseq, alt)
	case *MultOpNode:
		nodes := make([]Node, len(n.nodes))
		for i := 0; i < len(n.nodes); i++ {
			nodes[i] = t.GenerateTestNode(g, n.nodes[i])
		}
		return NewMultOpNode(g, n.operator, nodes)
	case *SelectorNode:
		gotoNode := t.GenerateTestNode(g, n.gotoNode)
		return NewSelectorNode(g, gotoNode, n.selectors)
	case *UnOpNode:
		node := t.GenerateTestNode(g, n.node)
		return NewUnOpNode(g, n.operator, node)
	case *ValueNode:
		return NewValueNode(g, n.value)
	case *VarNode:
		return NewVarNode(g, n.name)
	default:
		panic("Invalid node type for GenerateTestNode()")
	}
}

func (t *Testing) NewBinOpNode(operator tokens.Token, a Node, b Node) Node {
	return &BinOpNode{
		id:       -1,
		operator: operator,
		a:        a,
		b:        b,
	}
}

func (t *Testing) NewForNode(name string, collection Node, body Node) Node {
	return &ForNode{
		id:         -1,
		name:       name,
		collection: collection,
		body:       body,
	}
}

func (t *Testing) NewGotoNode(url Node) Node {
	return &GotoNode{
		id:  -1,
		url: url,
	}
}

func (t *Testing) NewIfNode(pred Node, conseq Node, alt Node) Node {
	return &IfNode{
		id:     -1,
		pred:   pred,
		conseq: conseq,
		alt:    alt,
	}
}

func (t *Testing) NewMultOpNode(operator tokens.Token, nodes []Node) Node {
	return &MultOpNode{
		id:       -1,
		operator: operator,
		nodes:    nodes,
	}
}

func (t *Testing) NewSelectorNode(gotoNode Node, selectors []Selector) Node {
	return &SelectorNode{
		id:        -1,
		selectors: selectors,
		gotoNode:  gotoNode,
	}
}

func (t *Testing) NewUnOpNode(operator tokens.Token, node Node) Node {
	return &UnOpNode{
		id:       -1,
		operator: operator,
		node:     node,
	}
}

func (t *Testing) NewValueNode(value interface{}) Node {
	return &ValueNode{id: -1, value: value}
}

func (t *Testing) NewVarNode(name string) Node {
	return &VarNode{id: -1, name: name}
}
