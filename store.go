package main

import (
	"io"
	"log"
	"os"
)

type PathTransformFunc func(string) string

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(path string) string {
	return path
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(path string, r io.Reader) error {
	path = s.PathTransformFunc(path)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	filename := "somefile"
	pathAndFilename := path + "/" + filename

	f, err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("wrote %d bytes to %s", n, pathAndFilename)
	return nil
}
