package graph

import (
	"reflect"
	"testing"
)

func TestPopIndex(t *testing.T) {
	index := NewStreamIndex(3, 2)
	expected := StreamIndex{[]int{3, 2}, "32"}
	if !reflect.DeepEqual(index, expected) {
		t.Errorf("Different values: expected %v got %v", expected, index)
	}

	i, restIdx := index.PopIndex()
	if i != 2 {
		t.Errorf("Different values: expected %d got %d", 2, i)
	}

	if !reflect.DeepEqual(index, expected) {
		t.Errorf("Different values: expected %v got %v", expected, index)
	}

	expected = StreamIndex{[]int{3}, "3"}
	if !reflect.DeepEqual(restIdx, expected) {
		t.Errorf("Different values: expected %v got %v", expected, index)
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
		t.Errorf("Different nodes: expected %#v got %#v", expected, node)
	}
}

var getSetDataNodeTests = []struct {
	lens  []int
	index []int
	value any
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
		if err := node.Set(index, test.value); err != nil {
			t.Error(err)
		}

		result, err := node.Get(index)
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
