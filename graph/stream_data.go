package graph

import (
	"bytes"
	"errors"
	"fmt"
)

type StreamIndex struct {
	// Index from most nested nodes to least nested nodes.
	// For example [3 1] is the 1st index of the first collection
	// and the 3rd index of the second collection.
	Indexes []int
	s       string
}

func NewStreamIndex(idxs ...int) *StreamIndex {
	var strIdx bytes.Buffer
	for _, i := range idxs {
		strIdx.WriteString(fmt.Sprintf("%d", i))
	}

	return &StreamIndex{idxs, strIdx.String()}
}

func (s *StreamIndex) AddIndex(index int) {
	s.s += fmt.Sprintf("%d", index)
	s.Indexes = append(s.Indexes, index)
}

// Append adds the parameter index to the beginning of the index
// (the end of the array).
// Example: appending [1 2] to [3 4] yields [3 4 1 2]
func (s *StreamIndex) Append(appendIdx *StreamIndex) {
	s.s += appendIdx.s
	s.Indexes = append(s.Indexes, appendIdx.Indexes...)
}

// String returns the stream index as an indexable string format.
// Example: [1 3 4] returns "134".
func (s *StreamIndex) String() string {
	return s.s
}

func (s *StreamIndex) Clone() *StreamIndex {
	copiedIdxs := make([]int, len(s.Indexes))
	for i := 0; i < len(s.Indexes); i++ {
		copiedIdxs[i] = s.Indexes[i]
	}
	return &StreamIndex{copiedIdxs, s.s}
}

func (s *StreamIndex) PopIndex() int {
	if s.Empty() {
		return -1
	}
	poppedIdx := s.Indexes[len(s.Indexes)-1]
	s.Indexes = s.Indexes[:len(s.Indexes)-1]
	s.s = s.s[:len(s.s)-1]
	return poppedIdx
}

func (s *StreamIndex) PeekIndex() int {
	if s.Empty() {
		return -1
	}

	return s.Indexes[len(s.Indexes)-1]
}

func (i *StreamIndex) Empty() bool {
	return len(i.Indexes) == 0
}

func (i *StreamIndex) Len() int {
	if len(i.Indexes) <= 0 {
		return 0
	}

	l := i.Indexes[0]
	for idx := 1; idx < len(i.Indexes); idx++ {
		l *= i.Indexes[idx]
	}
	return l
}

type DataNode interface {
	// Set a value at the index.
	Set(idx *StreamIndex, data interface{}) error
	// Get a value at the index.
	Get(idx *StreamIndex) (interface{}, error)
	// Check if the subnode at the top level index is full.
	Full(idx *StreamIndex) bool
	// Return the data representation of the node.
	Data() interface{}
}

// NewDataNode creates a new data given a stream index of
// lengths of the nested collected data. For example [3 5]
// means that the first collection has length 5 and the second
// nested collection is of length 3, meaning there are 5 * 3 = 15
// slots total.
func NewDataNode(lens *StreamIndex) DataNode {
	clonedLens := lens.Clone()
	l := clonedLens.PopIndex()
	data := make([]DataNode, l)
	if !clonedLens.Empty() {
		for i := 0; i < l; i++ {
			data[i] = NewDataNode(clonedLens.Clone())
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
	clonedIdx := idx.Clone()
	i := clonedIdx.PopIndex()
	if i >= len(s.data) || i < 0 {
		return errors.New(fmt.Sprintf("Set index out of bounds: index: %d, length: %d", i, len(s.data)))
	}
	return s.data[i].Set(clonedIdx, data)
}

func (s *streamNode) Get(idx *StreamIndex) (interface{}, error) {
	clonedIdx := idx.Clone()
	i := clonedIdx.PopIndex()
	if i >= len(s.data) || i < 0 {
		return nil, errors.New(fmt.Sprintf("Get index out of bounds: index: %d, length: %d", i, len(s.data)))
	}
	return s.data[i].Get(clonedIdx)
}

func (s *streamNode) Full(idx *StreamIndex) bool {
	if idx.Empty() {
		isFull := true
		for _, node := range s.data {
			if !node.Full(idx) {
				isFull = false
			}
		}
		return isFull
	}

	clonedIdx := idx.Clone()
	i := clonedIdx.PopIndex()
	return s.data[i].Full(clonedIdx)
}

func (s *streamNode) Data() interface{} {
	if len(s.data) == 0 {
		return nil
	} else if len(s.data) == 1 {
		return s.data[0].Data()
	}

	results := make([]interface{}, len(s.data))
	for i, data := range s.data {
		results[i] = data.Data()
	}
	return results
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

func (s *streamLeaf) Full(idx *StreamIndex) bool {
	return s.data != nil
}

func (s *streamLeaf) Data() interface{} {
	return s.data
}
