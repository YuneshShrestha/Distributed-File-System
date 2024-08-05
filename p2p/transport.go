package p2p

import "net"

// Peer is an interface that represents a remote node in the network.
type Peer interface {
	Send([]byte) error
	RemoteAddr() net.Addr
	Close() error
}

// Transport is an interface that handles the communication between peers.
// This can be a TCP connection, a UDp connection , a Websocket connection, etc.
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Dial(addr string) error
}
