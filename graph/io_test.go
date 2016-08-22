package graph

import (
	"bytes"
	"encoding/gob"
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
