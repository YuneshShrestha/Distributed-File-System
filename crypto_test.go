package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncrptyAndDecrypt(t *testing.T) {
	src := bytes.NewReader([]byte("Hello, World!"))
	dst := new(bytes.Buffer)

	key := newEncryptionKey()
	_, err := copyEncrypt(dst, src, key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dst.String())
	out := new(bytes.Buffer)
	_, err = copyDecrypt(out, dst, key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out.String())
}
