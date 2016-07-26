package graph

import (
	"errors"
	"fmt"
)

type StreamIndex struct {
	// Index from most nested nodes to least nested nodes.
	// For example [3 1] is the 1st index of the first collection
	// and the 3rd index of the second collection.
	Indexes []int
}

func NewStreamIndex(idxs ...int) *StreamIndex {
	return &StreamIndex{idxs}
}

func (s *StreamIndex) AddIndex(index int) {
	s.Indexes = append(s.Indexes, index)
}

func (s *StreamIndex) Clone() *StreamIndex {
	copiedIdxs := make([]int, len(s.Indexes))
	for i := 0; i < len(s.Indexes); i++ {
		copiedIdxs[i] = s.Indexes[i]
	}
	return &StreamIndex{copiedIdxs}
}

func (s *StreamIndex) PopIndex() int {
	if s.Empty() {
		return -1
	}
	poppedIdx := s.Indexes[len(s.Indexes)-1]
	s.Indexes = s.Indexes[:len(s.Indexes)-1]
	return poppedIdx
}

func (i *StreamIndex) Empty() bool {
	return len(i.Indexes) == 0
}

type DataNode interface {
	Set(idx *StreamIndex, data interface{}) error
	Get(idx *StreamIndex) (interface{}, error)
}

// NewDataNode creates a new data given a stream index of
// lengths of the nested collected data. For example [3 5]
// means that the first collection has length 5 and the second
// nested collection is of length 3, meaning there are 5 * 3 = 15
// slots total.
func NewDataNode(lens *StreamIndex) DataNode {
	l := lens.PopIndex()
	data := make([]DataNode, l)
	if !lens.Empty() {
		for i := 0; i < l; i++ {
			data[i] = NewDataNode(lens.Clone())
		}
	} else {
		for i := 0; i < l; i++ {
			data[i] = &streamLeaf{nil}
		}
	}
	return &streamNode{data}
}

type streamNode struct {
	data []DataNode
}

func (s *streamNode) Set(idx *StreamIndex, data interface{}) error {
	i := idx.PopIndex()
	if i >= len(s.data) || i < 0 {
		return errors.New(fmt.Sprintf("Set index out of bounds: index: %d, length: %d", i, len(s.data)))
	}
	return s.data[i].Set(idx, data)
}

func (s *streamNode) Get(idx *StreamIndex) (interface{}, error) {
	i := idx.PopIndex()
	if i >= len(s.data) || i < 0 {
		return nil, errors.New(fmt.Sprintf("Get index out of bounds: index: %d, length: %d", i, len(s.data)))
	}
	return s.data[i].Get(idx)
}

type streamLeaf struct {
	data interface{}
}

func (s *streamLeaf) Set(idx *StreamIndex, data interface{}) error {
	if idx.Empty() {
		s.data = data
		return nil
	}
	return errors.New("Setting data with non-empty index")
}

func (s *streamLeaf) Get(idx *StreamIndex) (interface{}, error) {
	if idx.Empty() {
		return s.data, nil
	}
	return nil, errors.New("Getting data with non-empty index")
}
