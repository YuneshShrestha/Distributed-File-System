package main

import (
	"bytes"
	"fmt"
	"io"
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
		EncryptionKey:     newEncryptionKey(),
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
	key := "mypics"

	// for i := 0; i < 10; i++ {
	data := bytes.NewReader([]byte("Hello"))

	server2.Store(key, data)

	// }
	// This part of the code is performing the following actions:
	time.Sleep(3 * time.Second)
	if server2.store.Delete(key) != nil {
		log.Fatal("Failed to delete key")

	}

	// // Delete key from server2
	if err := server2.store.Delete(key); err != nil {
		log.Fatal(err)
	}
	r, err := server2.Get(key)
	if err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	fmt.Println(buf.String())

	select {}
}
