package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

func makeServer(listenAddr string, root string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		Transport:         tcpTransport,
		StorageRoot:       root + "_network",
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

	server1 := makeServer(":3000", "3000")
	server2 := makeServer(":4000", "4000", ":3000")

	go func() {
		log.Fatal(server1.Start())
	}()
	time.Sleep(2 * time.Second)
	go server2.Start()
	time.Sleep(2 * time.Second)
	for i := 0; i < 10; i++ {
		data := bytes.NewReader([]byte("My data file here"))

		server2.Store(fmt.Sprintf("private%d", i), data)

		time.Sleep(1 * time.Second)
	}
	

	select {}
}
