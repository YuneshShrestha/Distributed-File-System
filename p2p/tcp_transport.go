package p2p

import (
	"errors"
	"fmt"

	"net"
	"sync"
)

// TCPPeer repeesents a remote node over a TCP established connection.
type TCPPeer struct {
	// the underlying TCP connection.
	net.Conn

	// if we are the dialer, outbound will be true.
	// if we are the listener, outbound will be false.
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// Send implements the Peer interface, which will send the message to the remote node.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(*TCPPeer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

// Close implements the Transport interface, which will close the underlying TCP connection.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
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

// Addr implements the Transport interface, which will return the address of the current node in the network.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Dial implements the Transport interface, which will dial a remote node in the network.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConnection(conn, true)
	return nil
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
		if errors.Is(err, net.ErrClosed) {
			fmt.Println("listener closed")

		}
		if err != nil {
			fmt.Printf("TCP accept error: %s \n", err)

		}
		fmt.Printf("new incomming connection: %s\n", conn.RemoteAddr())

		go t.handleConnection(conn, false)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		t.rpcch <- rpc
	}
}
