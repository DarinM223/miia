package graph

import (
	"encoding/gob"
	"io"
)

/*
 * Helper functions for reading and writing to files.
 */

func ReadInt(r io.Reader) (int, error) {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return -1, err
	}

	return int(buf[0] + (buf[1] << 8) + (buf[2] << 16) + (buf[3] << 24)), nil
}

func ReadString(r io.Reader) (string, error) {
	len, err := ReadInt(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, len)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}

	return string(buf[:]), nil
}

func ReadInterface(r io.Reader) (result interface{}, err error) {
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&result)
	return
}

func WriteInt(w io.Writer, i int) error {
	b3 := byte((i >> 24) & (0xFF))
	b2 := byte((i >> 16) & (0xFF))
	b1 := byte((i >> 8) & (0xFF))
	b0 := byte(i & 0xFF)

	_, err := w.Write([]byte{b0, b1, b2, b3})
	return err
}

func WriteString(w io.Writer, s string) error {
	if err := WriteInt(w, len(s)); err != nil {
		return err
	}

	bytes := []byte(s)
	_, err := w.Write(bytes)
	return err
}

func WriteInterface(w io.Writer, i interface{}) error {
	encoder := gob.NewEncoder(w)
	return encoder.Encode(&i)
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
