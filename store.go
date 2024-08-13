package main

import (
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
	// check if file exists
	return err == nil
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
	fmt.Printf("Deleting %s\n", firstPathNameWithRoot)
	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}
func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int, error) {
	// PathTransformFunc is a function that takes a key and returns a PathKey
	f, err := s.openFileForWriting(key)

	defer func() {

		f.Close()
	}()
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(f, r, encKey)

	return int(n), err
}
func (s *Store) openFileForWriting(key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.Pathname)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	fullPathWithRoot := fmt.Sprintf("%s/%s", pathNameWithRoot, pathKey.Filename)
	return os.Create(fullPathWithRoot)
}
func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	// PathTransformFunc is a function that takes a key and returns a PathKey
	f, err := s.openFileForWriting(key)
	defer func() {

		f.Close()
	}()
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)

}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}
