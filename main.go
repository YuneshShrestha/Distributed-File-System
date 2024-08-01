package main

import (
	"log"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":4000",
		HandshakeFunc: p2p.NOPHandshakeFunc,

		Decoder: p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts) // tr is a pointer to a TCPTransport struct

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

}
