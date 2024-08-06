package p2p



// RPC holds any arbitrary data that needs to be sent between peers over each transport.
type RPC struct {
	From string
	Payload []byte
}