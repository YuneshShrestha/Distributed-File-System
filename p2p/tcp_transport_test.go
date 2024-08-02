package p2p

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAddr := ":4000"
	tcpOpts := TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(tcpOpts)
	err := tr.ListenAndAccept()
	// check if the listen address is correct
	assert.Equal(t, listenAddr, tr.ListenAddr)

	assert.IsType(t, &net.OpError{}, err)

}
