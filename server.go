package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string

	// TCPTransportOpts  p2p.TCPTransportOpts
}

type FileServer struct {
	FileServerOpts

	store    *Store
	quitch   chan struct{} // channel to signal the server to stop: quit channel
	peers    map[string]p2p.Peer
	peerLock sync.Mutex
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p
	fmt.Println("Peer connected: ", p.RemoteAddr())

	return nil

}

func (s *FileServer) loop() {
	defer func() {
		log.Println("Closing the server")
		s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println("Received message: ", msg)
		case <-s.quitch:
			return

		}
	}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Payload struct {
	Key  string
	Data []byte
}

func (s *FileServer) boardCast(p Payload) error {
	peers := make([]io.Writer, 0)

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}
func (s *FileServer) StoreData(key string, r io.Reader) error {

	if err := s.store.Write(key, r); err != nil {
		fmt.Println("Error storing data: ", err)
		return err
	}
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		fmt.Println("Error copying data to buffer: ", err)
		return err
	}
	p := Payload{
		Key:  key,
		Data: buf.Bytes(),
	}
	fmt.Println("Storing data: ", buf.Bytes())
	return s.boardCast(p)

}
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func() {
			fmt.Println("Dialing ", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("Failed to dial %s: %s", addr, err)
			}
		}()
	}
	return nil
}
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if err := s.bootstrapNetwork(); err != nil {
		return err

	}

	s.loop()
	return nil
}
