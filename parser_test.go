package main

import (
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
		"num1+num2", "num1",
		0, 4,
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
		err := parser.expectString(test.expected)
		if err != test.err {
			t.Errorf("Different errors: expected %v got %v", test.err, err)
		}
		if parser.pos != test.endPos {
			t.Errorf("Different end positions: expected %d got %d", test.endPos, parser.pos)
		}
	}
}
