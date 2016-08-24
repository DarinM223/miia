package graph

import (
	"bytes"
	"encoding/gob"
	"github.com/DarinM223/miia/tokens"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
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
		t.Errorf("Expected %d got %d", str, result)
	}
}

func TestReadWriteInterface(t *testing.T) {
	var buf bytes.Buffer

	gob.Register([]interface{}{})
	gob.Register(map[int]string{})

	obj := []interface{}{[]interface{}{1, 2, 3}, map[int]string{1: "Hello", 2: "World"}, 69}
	WriteInterface(&buf, obj)

	result, err := ReadInterface(&buf)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(result, obj) {
		t.Errorf("Expected %v got %v", obj, result)
	}
}

var nodeReadWriteTests = []Node{
	testUtils.NewBinOpNode(tokens.AndToken, testUtils.NewValueNode(true), testUtils.NewValueNode(true)),
	testUtils.NewCollectNode(testUtils.NewValueNode(1)),
	testUtils.NewValueNode([]int{1, 2, 3}),
	testUtils.NewVarNode("a"),
	testUtils.NewForNode("a", testUtils.NewValueNode(1), testUtils.NewVarNode("a")),
	testUtils.NewCollectNode(testUtils.NewForNode("a", testUtils.NewValueNode([]int{1, 2, 3}), testUtils.NewVarNode("a"))),
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
	for _, test := range nodeReadWriteTests {
		var buf bytes.Buffer

		g := NewGlobals()
		node := testUtils.GenerateTestNode(g, test)
		node.Write(&buf)

		readNode, err := ReadNode(&buf, NewGlobals())
		if err != nil {
			t.Error(err)
		}

		if !testUtils.CompareTestNodeToRealNode(readNode, node) {
			t.Errorf("Expected %v got %v", spew.Sdump(node), spew.Sdump(readNode))
		}
	}
}
