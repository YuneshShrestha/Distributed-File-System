package main

import (
	"bytes"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("Hello World"))
	err := s.writeStream("somepath", data)
	if err != nil {
		t.Fatal(err)
	}
}
