package main

import (
	"github.com/DarinM223/http-scraper/graph"
	"testing"
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

func TestCompileSelector(t *testing.T) {
	// TODO(DarinM223): implement this
}

func TestCompileVar(t *testing.T) {
	// TODO(DarinM223): implement this
}
