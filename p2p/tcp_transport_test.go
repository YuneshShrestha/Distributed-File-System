package p2p

import (
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

	// check if the listen address is correct
	assert.Equal(t, listenAddr, tr.ListenAddr)

	assert.Nil(t, tr.ListenAndAccept())

}
