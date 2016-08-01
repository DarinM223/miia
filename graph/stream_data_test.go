package graph

import (
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

func TestPopIndex(t *testing.T) {
	index := NewStreamIndex(3, 2)
	expected := &StreamIndex{[]int{3, 2}, "32"}
	if !reflect.DeepEqual(index, expected) {
		t.Errorf("Different values: expected %v got %v", expected, index)
	}

	i := index.PopIndex()
	if i != 2 {
		t.Errorf("Different values: expected %d got %d", 2, i)
	}

	expected = &StreamIndex{[]int{3}, "3"}
	if !reflect.DeepEqual(index, expected) {
		t.Errorf("Different values: expected %v got %v", expected, index)
	}
}

func TestCloneIndex(t *testing.T) {
	index := NewStreamIndex(3, 2)
	clone := index.Clone()

	index.AddIndex(3)

	expectedIndex := &StreamIndex{[]int{3, 2, 3}, "323"}
	expectedClone := &StreamIndex{[]int{3, 2}, "32"}

	if !reflect.DeepEqual(index, expectedIndex) {
		t.Errorf("Different values: expected %v got %v", expectedIndex, index)
	}

	if !reflect.DeepEqual(clone, expectedClone) {
		t.Errorf("Different values: expected %v got %v", expectedClone, clone)
	}
}

func TestNewDataNode(t *testing.T) {
	lens := NewStreamIndex(3, 2)
	node := NewDataNode(lens)
	expected := &streamNode{
		[]DataNode{
			&streamNode{
				[]DataNode{&streamLeaf{nil}, &streamLeaf{nil}, &streamLeaf{nil}},
			},
			&streamNode{
				[]DataNode{&streamLeaf{nil}, &streamLeaf{nil}, &streamLeaf{nil}},
			},
		},
	}

	if !reflect.DeepEqual(node, expected) {
		t.Errorf("Different nodes: expected %v got %v", spew.Sdump(expected), spew.Sdump(node))
	}
}

var getSetDataNodeTests = []struct {
	lens  []int
	index []int
	value interface{}
}{
	{
		[]int{3, 2},
		[]int{2, 1},
		2,
	},
	{
		[]int{3, 2},
		[]int{0, 0},
		"hello",
	},
	{
		[]int{1},
		[]int{0},
		"hello",
	},
}

func TestGetDataNode(t *testing.T) {
	for _, test := range getSetDataNodeTests {
		lens := NewStreamIndex(test.lens...)
		index := NewStreamIndex(test.index...)

		node := NewDataNode(lens)
		if err := node.Set(index.Clone(), test.value); err != nil {
			t.Error(err)
		}

		result, err := node.Get(index.Clone())
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(result, test.value) {
			t.Errorf("Different values: expected %v got %gv", test.value, result)
		}
	}
}

var dataNodeLenTests = []struct {
	lens   []int
	length int
}{
	{
		[]int{1, 2, 3},
		6,
	},
	{
		[]int{6},
		6,
	},
	{
		[]int{2, 3},
		6,
	},
}

func TestDataNodeLen(t *testing.T) {
	for _, test := range dataNodeLenTests {
		lens := NewStreamIndex(test.lens...)

		if lens.Len() != test.length {
			t.Errorf("Different lengths: expected %v got %v", test.length, lens.Len())
		}
	}
}
