package p2p

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

// CloseStream implements the Peer interface, which will close the underlying TCP connection.
func (p *TCPPeer) CloseStream() error {
	return p.Conn.Close()
}

// Send implements the Peer interface, which will send a message to the remote peer.
func (p *TCPPeer) Send(data []byte) error {

	// Create a buffer to hold the length prefix and the data
	var buf bytes.Buffer

	// Write the length of the data as a 4-byte integer
	length := int32(len(data))
	if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
		return err
	}

	// Write the actual data
	if _, err := buf.Write(data); err != nil {
		return err
	}

	// Send the buffer contents
	_, err := p.Conn.Write(buf.Bytes())
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

// Addr implements the Transport interface, which will return the address of the listener.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Consume implements the Transport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
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
		fmt.Printf("closing connection: %s\n", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)

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
		// Read the length of the incoming message
		var length int32
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			if err == io.EOF {
				
				fmt.Printf("peer closed the connection: %s\n", conn.RemoteAddr())
				return
			}
			fmt.Printf("failed to read message length: %s\n", err)
			return
		}

		// Read the actual message based on the length
		message := make([]byte, length)
		if _, err := io.ReadFull(reader, message); err != nil {
			fmt.Printf("failed to read message: %s\n", err)
			return
		}

		lineReader := bytes.NewReader(message)

		if err := t.Decoder.Decode(lineReader, &rpc); err != nil {
			fmt.Printf("failed to decode message: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr().String()
		peer.wg.Add(1)
		fmt.Println("Waiting for the peer to read the message")
		t.rpcch <- rpc

		peer.wg.Wait()
		fmt.Println("Stream Done continue normal reading")
	}
}
