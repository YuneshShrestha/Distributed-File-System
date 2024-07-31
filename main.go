package main

import (
	"log"

	"github.com/YuneshShrestha/Distributed-File-System-Codes/p2p"
)

func main() {
	tr := p2p.NewTCPTransport(":4000")
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to listen and accept: %v", err)
	}

	select {}

}
