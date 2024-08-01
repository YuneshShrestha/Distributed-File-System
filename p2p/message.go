package p2p

import "net"

// Message holds any arbitrary data that needs to be sent between peers over each transport.
type Message struct {
	From net.Addr
	Payload []byte
}