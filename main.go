package main

import (
	"fmt"
	"log"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

func OnPeer(peer p2p.Peer) error {
	// peer.Close()
	return nil
}
func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":4000",
		HandshakeFunc: p2p.NOPHandshakeFunc,

		Decoder: p2p.DefaultDecoder{},
		OnPeer:  OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts) // tr is a pointer to a TCPTransport struct

	// go routine to consume the messages from the channel
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

}
