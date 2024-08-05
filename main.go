package main

import (
	"log"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		Transport:         tcpTransport,
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = func(p *p2p.TCPPeer) error {
		return s.OnPeer(p)
	}
	return s

}

func main() {
	server1 := makeServer(":8080")
	server2 := makeServer(":8081", ":8080")

	go func() {
		log.Fatal(server1.Start())
	}()

	server2.Start()
}
