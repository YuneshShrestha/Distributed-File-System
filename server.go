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
		case rpc := <-s.Transport.Consume():
			// Received payload from peer

			var msg Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("Failed to decode message: ", err)
				continue
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Failed to handle message: ", err)
				continue
			}
			// fmt.Printf("Received message: %+v\n", msg)

			// peer, ok := s.peers[rpc.From]
			// if !ok {
			// 	log.Println("Peer not found: ", rpc.From)
			// 	continue
			// }

			// b := make([]byte, 1000)
			// n, err := peer.Read(b)
			// if err != nil {
			// 	log.Println("Failed to read from peer: ", err)
			// 	continue
			// }
			// fmt.Println("Received data: ", string(b[:n]))
			// peer.(*p2p.TCPPeer).Wg.Done()

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
	// From    string
	Payload any
}
type MessageStoreFile struct {
	Key  string
	Size int
}

func init() {
	gob.Register(MessageStoreFile{})
}

// type DataMessage struct {
// 	Key  string
// 	Data []byte
// }

func (s *FileServer) boardCast(p *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)
}
func (s *FileServer) StoreData(key string, r io.Reader) error {
	var (
		filebuf = new(bytes.Buffer)
		tee     = io.TeeReader(r, filebuf)
	)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		fmt.Println("Error encoding message: ", err)

		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	for _, peer := range s.peers {
		n, err := io.Copy(peer, filebuf)
		if err != nil {
			return err
		}
		fmt.Println("Sent data to peer: ", n)
	}
	return nil

}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}
	return nil

}
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer not found: %s", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size)))

	if err != nil {
		fmt.Println("Error writing to store: ", err)
		return err
	}
	fmt.Println("Received data from peer: ", n)
	peer.(*p2p.TCPPeer).Wg.Done()
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
