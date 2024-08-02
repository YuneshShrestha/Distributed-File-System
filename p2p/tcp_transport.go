package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

// TCPPeer repeesents a remote node over a TCP established connection.
type TCPPeer struct {
	// conn is the underlying TCP connection.
	conn net.Conn

	// if we are the dialer, outbound will be true.
	// if we are the listener, outbound will be false.
	outbound bool
}

func NewTCPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// Close implements the Peer interface, which will close the underlying TCP connection.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the Transport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	fmt.Printf("listening on %s\n", t.ListenAddr)
	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s \n", err)
			continue
		}
		go t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("closing connection: %s\n", err)
		conn.Close()
	}()
	peer := NewTCPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("failed to shake hands: %s\n", err)
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			fmt.Printf("failed to handle peer: %s\n", err)
			return
		}
	}

	rpc := RPC{}
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {

			if err == io.EOF {
				fmt.Printf("peer closed the connection: %s\n", conn.RemoteAddr())
				return
			}
			if opErr, ok := err.(*net.OpError); ok {
				fmt.Printf("network operation error: %s\n", opErr)
				// You can add more specific handling for *net.OpError here if needed
			} else {
				fmt.Printf("failed to read message: %s\n", err)
			}
			return
		}
		lineReader := bytes.NewReader([]byte(line))

		if err := t.Decoder.Decode(lineReader, &rpc); err != nil {
			fmt.Printf("failed to decode message: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
