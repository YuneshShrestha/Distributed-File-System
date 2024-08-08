package p2p


const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)
// RPC holds any arbitrary data that needs to be sent between peers over each transport.
type RPC struct {
	From string
	Payload []byte
	Stream bool
}