package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

// CAS = Content Addressable Storage is a method of storing data such that it is retrievable based on its content, not its location.
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		panic("Pathname is empty")
	}
	return paths[0]
}
func (p PathKey) FullPath() string {
	return p.Pathname + "/" + p.Filename
}

type StoreOpts struct {
	// Root is the root directory where the files/folders will be stored
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return err != os.ErrNotExist
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}
func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("Deleted %s", pathKey.Filename)
	}()
	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())
	return os.RemoveAll(firstPathNameWithRoot)
}
func (s *Store) writeStream(key string, r io.Reader) (int, error) {
	// PathTransformFunc is a function that takes a key and returns a PathKey
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.Pathname)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}
	fullPathWithRoot := fmt.Sprintf("%s/%s", pathNameWithRoot, pathKey.Filename)
	f, err := os.Create(fullPathWithRoot)
	defer func() {

		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (s *Store) Write(key string, r io.Reader) (int, error) {
	return s.writeStream(key, r)
}
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()
	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return buf, nil

}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(fullPathWithRoot)
}
