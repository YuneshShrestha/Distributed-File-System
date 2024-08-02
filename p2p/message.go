package p2p

import "net"

// RPC holds any arbitrary data that needs to be sent between peers over each transport.
type RPC struct {
	From net.Addr
	Payload []byte
}