package main

import (
	"log"
	"time"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":4000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		Transport:         tcpTransport,
		StorageRoot:       "4000_root",
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewFileServer(fileServerOpts)
	go func() {
		time.Sleep(5 * time.Second)
		s.Stop()
	}()
	if err := s.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
