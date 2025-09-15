package graph

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/DarinM223/miia/tokens"
	"github.com/davecgh/go-spew/spew"
)

func TestReadWriteInt(t *testing.T) {
	var buf bytes.Buffer

	num := 69
	if err := WriteInt(&buf, num); err != nil {
		t.Error(err)
	}

	result, err := ReadInt(&buf)
	if err != nil {
		t.Error(err)
	}

	if result != num {
		t.Errorf("Expected %d got %d", num, result)
	}
}

func TestReadWriteInt64(t *testing.T) {
	var buf bytes.Buffer

	var num int64 = 1234567890
	if err := WriteInt64(&buf, num); err != nil {
		t.Error(err)
	}

	result, err := ReadInt64(&buf)
	if err != nil {
		t.Error(err)
	}

	if result != num {
		t.Errorf("Expected %d got %d", num, result)
	}
}

func TestReadWriteString(t *testing.T) {
	var buf bytes.Buffer

	str := "Hello world!"
	if err := WriteString(&buf, str); err != nil {
		t.Error(err)
	}

	result, err := ReadString(&buf)
	if err != nil {
		t.Error(err)
	}

	if result != str {
		t.Errorf("Expected %s got %s", str, result)
	}
}

var readWriteValueTests = []any{
	123456789,
	true,
	"hello",
}

func TestReadWriteValue(t *testing.T) {
	for _, test := range readWriteValueTests {
		var buf bytes.Buffer
		err := WriteValue(&buf, test)
		if err != nil {
			t.Error(err)
		}

		result, err := ReadValue(&buf)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(result, test) {
			t.Errorf("Expected %v got %v", test, result)
		}
	}
}

var nodeReadWriteTests = []Node{
	testUtils.NewBinOpNode(tokens.AndToken, testUtils.NewValueNode(true), testUtils.NewValueNode(true)),
	testUtils.NewCollectNode(testUtils.NewValueNode(1)),
	testUtils.NewValueNode(true),
	testUtils.NewValueNode(100),
	testUtils.NewValueNode("http://example.com"),
	testUtils.NewValueNode(nil),
	testUtils.NewVarNode("a"),
	testUtils.NewForNode("a", testUtils.NewValueNode(1), testUtils.NewVarNode("a")),
	testUtils.NewCollectNode(
		testUtils.NewForNode(
			"a",
			testUtils.NewBinOpNode(tokens.RangeToken, testUtils.NewValueNode(1), testUtils.NewValueNode(10)),
			testUtils.NewVarNode("a"),
		),
	),
	testUtils.NewGotoNode(testUtils.NewValueNode("http://www.google.com")),
	testUtils.NewIfNode(
		testUtils.NewValueNode(true),
		testUtils.NewValueNode(1),
		testUtils.NewValueNode(2),
	),
	testUtils.NewMultOpNode(
		tokens.AddToken,
		[]Node{
			testUtils.NewValueNode(1),
			testUtils.NewValueNode(2),
			testUtils.NewValueNode(3),
		},
	),
	testUtils.NewSelectorNode(
		testUtils.NewGotoNode(testUtils.NewValueNode("http://www.google.com")),
		[]Selector{
			{"blah", "p"},
			{"foo", "h2"},
		},
	),
	testUtils.NewUnOpNode(tokens.NotToken, testUtils.NewValueNode(true)),
}

func TestReadWriteNode(t *testing.T) {
	for i, test := range nodeReadWriteTests {
		var buf bytes.Buffer

		g := NewGlobals()
		node := testUtils.GenerateTestNode(g, test)
		node.Write(&buf)

		readNode, err := ReadNode(&buf, NewGlobals())
		if err != nil {
			t.Errorf("Error with test %d: %v", i, err)
		}

		if !testUtils.CompareTestNodeToRealNode(readNode, node) {
			t.Errorf("Expected %v got %v", spew.Sdump(node), spew.Sdump(readNode))
		}
	}
}

func TestReadWriteGlobals(t *testing.T) {
	var buf bytes.Buffer
	g := NewGlobals()
	valNode1 := NewValueNode(g, g.GenID(), true)
	valNode2 := NewValueNode(g, g.GenID(), true)
	resultNode := NewBinOpNode(g, g.GenID(), tokens.AndToken, valNode1, valNode2)
	g.SetResultNodeID(resultNode.ID())
	WriteGlobals(&buf, g)

	ng, err := ReadGlobals(&buf)
	if err != nil {
		t.Error(err)
	}

	if ng.currID != g.currID {
		t.Errorf("CurrID different: expected %d got %d", g.currID, ng.currID)
	}
	if ng.resultID != g.resultID {
		t.Errorf("ResultID different: expected %d got %d", g.resultID, ng.resultID)
	}
	if !reflect.DeepEqual(ng.rateLimiterData, g.rateLimiterData) {
		t.Errorf("Rate limiter data different: expected %v got %v", g.rateLimiterData, ng.rateLimiterData)
	}
	if !testUtils.CompareTestNodeToRealNode(g.ResultNode(), ng.ResultNode()) {
		t.Errorf("Result nodes different: expected %v got %v", spew.Sdump(g.ResultNode()), spew.Sdump(ng.ResultNode()))
	}
}
