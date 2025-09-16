package main

import (
	"testing"

	"github.com/DarinM223/miia/graph"
	"github.com/DarinM223/miia/tokens"
)

var graphTestUtils = graph.Testing{}

var compilerTests = []struct {
	expr     Expr
	expected graph.Node
}{
	{
		IntExpr{2},
		graphTestUtils.NewValueNode(2),
	},
	{
		StringExpr{"hello"},
		graphTestUtils.NewValueNode("hello"),
	},
	{
		IfExpr{
			BoolExpr{true},
			IntExpr{1},
			IntExpr{0},
		},
		graphTestUtils.NewIfNode(
			graphTestUtils.NewValueNode(true),
			graphTestUtils.NewValueNode(1),
			graphTestUtils.NewValueNode(0),
		),
	},
	{
		BlockExpr{
			[]Expr{
				BindExpr{map[string]Expr{"url": StringExpr{"http://www.google.com"}}},
				VarExpr{"url"},
			},
		},
		graphTestUtils.NewValueNode("http://www.google.com"),
	},
	{
		GotoExpr{StringExpr{"http://www.google.com"}},
		graphTestUtils.NewGotoNode(graphTestUtils.NewValueNode("http://www.google.com")),
	},
	{
		BlockExpr{
			[]Expr{
				GotoExpr{StringExpr{"http://www.google.com"}},
				SelectorExpr{[]graph.Selector{{"a", "b"}, {"c", "d"}}},
			},
		},
		graphTestUtils.NewSelectorNode(
			graphTestUtils.NewGotoNode(graphTestUtils.NewValueNode("http://www.google.com")),
			[]graph.Selector{{"a", "b"}, {"c", "d"}},
		),
	},
	{
		UnOp{
			tokens.NotToken,
			BoolExpr{true},
		},
		graphTestUtils.NewUnOpNode(tokens.NotToken, graphTestUtils.NewValueNode(true)),
	},
	{
		BinOp{
			tokens.AndToken,
			BoolExpr{true},
			BoolExpr{false},
		},
		graphTestUtils.NewBinOpNode(
			tokens.AndToken,
			graphTestUtils.NewValueNode(true),
			graphTestUtils.NewValueNode(false),
		),
	},
	{
		MultOp{
			tokens.AddToken,
			[]Expr{
				IntExpr{2},
				IntExpr{3},
				IntExpr{4},
			},
		},
		graphTestUtils.NewMultOpNode(
			tokens.AddToken,
			[]graph.Node{
				graphTestUtils.NewValueNode(2),
				graphTestUtils.NewValueNode(3),
				graphTestUtils.NewValueNode(4),
			},
		),
	},
	{
		CollectExpr{
			ForExpr{
				BinOp{tokens.RangeToken, IntExpr{1}, IntExpr{3}},
				"i",
				VarExpr{"i"},
			},
		},
		graphTestUtils.NewCollectNode(
			graphTestUtils.NewForNode(
				"i",
				graphTestUtils.NewBinOpNode(
					tokens.RangeToken,
					graphTestUtils.NewValueNode(1),
					graphTestUtils.NewValueNode(3),
				),
				graphTestUtils.NewVarNode("i"),
			),
		),
	},
}

func TestCompiler(t *testing.T) {
	for _, test := range compilerTests {
		globals := graph.NewGlobals()
		scope := NewScope(nil)
		node, err := CompileExpr(globals, test.expr, scope)
		if err != nil {
			t.Error(err)
		}
		if !graphTestUtils.CompareTestNodeToRealNode(test.expected, node) {
			t.Errorf("Different node result: expected %v got %v", test.expected, node)
		}
	}
}
