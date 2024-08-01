package p2p

// Message holds any arbitrary data that needs to be sent between peers over each transport.
type Message struct {
	Payload []byte
}