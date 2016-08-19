package graph

import (
	"encoding/gob"
	"io"
)

/*
 * Helper functions for reading and writing to files.
 */

func ReadInt(r io.Reader) int {
	// TODO(DarinM223): implement this
	return 0
}

func ReadString(r io.Reader) string {
	// TODO(DarinM223): implement this
	return ""
}

func ReadByte(r io.Reader) byte {
	// TODO(DarinM223): implement this
	return 0
}

func ReadInterface(r io.Reader) (result interface{}, err error) {
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(result)
	return
}

func WriteInt(w io.Writer, i int) {
	// TODO(DarinM223): implement this
}

func WriteString(w io.Writer, s string) {
	// TODO(DarinM223): implement this
}

func WriteByte(w io.Writer, b byte) {
	// TODO(DarinM223): implement this
}

func WriteInterface(w io.Writer, i interface{}) error {
	encoder := gob.NewEncoder(w)
	return encoder.Encode(i)
}

/*
 * Implementations for reading nodes from files.
 */

func ReadNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	// Read first byte and determine the type.
	return nil
}

func readBinOpNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readCollectNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readForNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readGotoNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readIfNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readMultOpNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readSelectorNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readUnOpNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readValueNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

func readVarNode(r io.Reader) Node {
	// TODO(DarinM223): implement this
	return nil
}

/*
 * Implementations for writing nodes to files.
 */

func (n *BinOpNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *CollectNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *ForNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *GotoNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *IfNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *MultOpNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *SelectorNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *UnOpNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *ValueNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}

func (n *VarNode) Write(w io.Writer) {
	// TODO(DarinM223): implement this
}
