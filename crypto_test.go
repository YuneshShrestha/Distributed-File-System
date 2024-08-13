package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncrptyAndDecrypt(t *testing.T) {
	payload := "Hello, World!"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)

	key := newEncryptionKey()
	_, err := copyEncrypt(dst, src, key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dst.String())
	out := new(bytes.Buffer)
	n, err := copyDecrypt(out, dst, key)
	if err != nil {
		t.Fatal(err)
	}
	// 16 bytes for the IV
	if n != int64(len(payload)) +16 {
		t.Fatalf("Expected %d, got %d", len(payload), n)
	}

	if out.String() != payload {
		t.Fatalf("Expected %s, got %s", payload, out.String())
	}
	fmt.Println(out.String() == payload)	

}