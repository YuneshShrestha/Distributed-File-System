package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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
		case rpc := <-s.Transport.Consume():
			// Received payload from peer

			var msg Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("Failed to decode message: ", err)
				continue
			}
			fmt.Println("Received message: ", string(msg.Payload.([]byte)))

			peer, ok := s.peers[rpc.From]
			if !ok {
				log.Println("Peer not found: ", rpc.From)
				continue
			}

			b := make([]byte, 1000)
			n, err := peer.Read(b)
			if err != nil {
				log.Println("Failed to read from peer: ", err)
				continue
			}
			fmt.Println("Received data: ", string(b[:n]))
			peer.(*p2p.TCPPeer).Wg.Done()

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

type Message struct {
	From    string
	Payload any
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (s *FileServer) boardCast(p *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)
}
func (s *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	msg := Message{
		Payload: []byte("storageKey"),
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {

		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	time.Sleep(3 * time.Second)
	payload := []byte("This is a test")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}
	return nil

}

func (s *FileServer) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case *DataMessage:
		fmt.Println("Received data message: ", v)
	}
	return nil

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
