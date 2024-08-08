package p2p

import "net"

// Peer is an interface that represents a remote node in the network.
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport is an interface that handles the communication between peers.
// This can be a TCP connection, a UDp connection , a Websocket connection, etc.
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Addr() string
	Dial(string) error
}
