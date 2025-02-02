package main

import (
	"bytes"
	"fmt"
	"io"

	// "io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	path := "somepaths"
	pathKey := CASPathTransformFunc(path)

	expectedPathName := "34c21/1d4b4/24a8b/a731a/0fdf5/ee47b/fcb4c/e99c1"
	expectedOriginalKey := "34c211d4b424a8ba731a0fdf5ee47bfcb4ce99c1"
	if pathKey.Pathname != expectedPathName {
		t.Fatalf("Expected %s, got %s", expectedPathName, pathKey.Pathname)
	}
	if pathKey.Filename != expectedOriginalKey {
		t.Fatalf("Expected %s, got %s", expectedOriginalKey, pathKey.Filename)
	}
}

func TestStoreDeleteKey(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	key := "myfavs"
	data := []byte("Hello Worlds")
	if _,err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}

	if !s.Has(key) {
		t.Fatal("Expected key to exist")
	}

	if err := s.Delete(key); err != nil {
		t.Fatal(err)
	}
}
func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key-%d", i)

		data := []byte("Hello Worlds")
		_,err := s.writeStream(key, bytes.NewReader(data))

		if err != nil {
			t.Fatal(err)
		}
		if ok := s.Has(key); !ok {
			t.Fatal("Expected key to exist")
		}
		_, r, err := s.Read(key)
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(data) {
			t.Fatalf("Expected %s, got %s", "Hello Worlds", string(b))
		}
		fmt.Println(string(b))

		if err := s.Delete(key); err != nil {
			t.Fatal(err)
		}

		if ok := s.Has(key); ok {
			t.Fatal("Expected key to not exist")
		}
	}

}
func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
