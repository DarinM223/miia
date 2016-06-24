package main

import (
	"reflect"
	"testing"
)

var parseIdentTests = []struct {
	text, ident string
	pos, endPos int
	err         error
}{
	{
		"hello1 ", "hello1",
		0, 6,
		nil,
	},
	{
		"1hello ", "",
		0, 0,
		NumFirstIdentErr,
	},
	{
		"hello", "hello",
		0, 5,
		nil,
	},
	{
		"+", "+",
		0, 1,
		nil,
	},
}

func TestParseIdent(t *testing.T) {
	for _, test := range parseIdentTests {
		parser := Parser{test.pos, test.text}
		ident, err := parser.parseIdent()
		if err != test.err {
			t.Errorf("Different errors: expected %v got %v", test.err, err)
		} else if ident != test.ident {
			t.Errorf("Different idents: expected %s got %s", test.ident, ident)
		}
		if parser.pos != test.endPos {
			t.Errorf("Different end positions: expected %d got %d", test.endPos, parser.pos)
		}
	}
}

var parseWhitspaceTests = []struct {
	text        string
	pos, endPos int
}{
	{
		"     hello",
		0, 5,
	},
	{
		"\t \n hello",
		0, 4,
	},
	{
		"hello   ",
		0, 0,
	},
	{
		"     ",
		0, 5,
	},
}

func TestParseWhitespace(t *testing.T) {
	for _, test := range parseWhitspaceTests {
		parser := Parser{test.pos, test.text}
		parser.parseWhitespace()

		if parser.pos != test.endPos {
			t.Errorf("Different end positions: expected %d got %d", test.endPos, parser.pos)
		}
	}
}

var expectStringTests = []struct {
	text, expected string
	pos, endPos    int
	err            error
}{
	{
		"abcdefg", "abcdefg",
		0, 7,
		nil,
	},
	{
		"hello", "h",
		0, 1,
		nil,
	},
	{
		"hello world", "world",
		6, 11,
		nil,
	},
	{
		"", "hello",
		0, 0,
		ExpectedStrErr,
	},
	{
		"hello world", "worlk",
		6, 6,
		ExpectedStrErr,
	},
}

func TestExpectString(t *testing.T) {
	for _, test := range expectStringTests {
		parser := Parser{test.pos, test.text}
		if err := parser.expectString(test.expected); err != test.err {
			t.Errorf("Different errors: expected %v got %v", test.err, err)
		}
		if parser.pos != test.endPos {
			t.Errorf("Different end positions: expected %d got %d", test.endPos, parser.pos)
		}
	}
}

var parseKeywordOrIndentTests = []struct {
	text, ident string
	token       Token
}{
	{"goto \"www.google.com\"", "goto", GotoToken},
	{"hello world", "hello", IdentToken},
	{"for name in names {", "for", ForToken},
	{"if (a == 2) {", "if", IfToken},
	{"else b", "else", ElseToken},
}

func TestParseKeywordOrIndent(t *testing.T) {
	for _, test := range parseKeywordOrIndentTests {
		parser := Parser{0, test.text}
		token, ident, err := parser.parseKeywordOrIdent()
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}

		if ident != test.ident {
			t.Errorf("Different idents: expected %s got %s", test.ident, ident)
		}
		if token != test.token {
			t.Errorf("Different tokens: expected %d got %d", test.token, token)
		}
	}
}

var parseNumberTests = []struct {
	text string
	expr Expr
	err  error
}{
	{"0", IntExpr{0}, nil},
	{"1234", IntExpr{1234}, nil},
	{"abcd", nil, NumErr},
	{"-23", IntExpr{-23}, nil},
}

func TestParseNumber(t *testing.T) {
	for _, test := range parseNumberTests {
		parser := Parser{0, test.text}
		expr, err := parser.parseNumber()
		if err != test.err {
			t.Errorf("Different errors: expected %v got %v", test.err, err)
		} else if err == nil && expr != test.expr {
			t.Errorf("Different exprs: expected %v got %v", test.expr, expr)
		}
	}
}

var parseStringTests = []struct {
	text string
	expr Expr
	err  error
}{
	{"\"Sample Text\"", StringExpr{"Sample Text"}, nil},
	{"\"Sample Text", nil, StringNotClosedErr},
	{"Sample Text\"", nil, StringNotClosedErr},
}

func TestParseString(t *testing.T) {
	for _, test := range parseStringTests {
		parser := Parser{0, test.text}
		expr, err := parser.parseString()
		if err != test.err {
			t.Errorf("Different errors: expected %v got %v", test.err, err)
		} else if err == nil && expr != test.expr {
			t.Errorf("Different exprs: expected %v got %v", test.expr, expr)
		}
	}
}

var parseExprTests = []struct {
	text string
	expr Expr
	err  error
}{
	{
		"(+ 1 2)",
		BinOp{AddToken, IntExpr{1}, IntExpr{2}},
		nil,
	},
	{
		"(+ a b)",
		BinOp{AddToken, VarExpr{"a"}, VarExpr{"b"}},
		nil,
	},
	{
		"(if 1 a \"Hello\")",
		IfExpr{IntExpr{1}, VarExpr{"a"}, StringExpr{"Hello"}},
		nil,
	},
	{
		"(block \"hello\" 2)",
		BlockExpr{[]Expr{StringExpr{"hello"}, IntExpr{2}}},
		nil,
	},
	{
		"(if (= a 1) (block (+ a b) \"hello\") 2)",
		IfExpr{
			BinOp{EqualsToken, VarExpr{"a"}, IntExpr{1}},
			BlockExpr{[]Expr{BinOp{AddToken, VarExpr{"a"}, VarExpr{"b"}}, StringExpr{"hello"}}},
			IntExpr{2},
		},
		nil,
	},
}

func TestParseExpr(t *testing.T) {
	for _, test := range parseExprTests {
		parser := Parser{0, test.text}
		expr, err := parser.parseExpr()
		if err != test.err {
			t.Errorf("Different errors: expected %v got %v", test.err, err)
		} else if err == nil && !reflect.DeepEqual(expr, test.expr) {
			t.Errorf("Different exprs: expected %v got %v", test.expr, expr)
		}
	}
}
