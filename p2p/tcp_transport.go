package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"sync"
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

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener   net.Listener
	shakeHands HandshakeFunc
	// decoder 	 Decoder
	mu    sync.RWMutex // RWMutex is a reader/writer mutual exclusion lock.  Mutex will protect the map of peers so they are witten above the field (in our case peers) that they are protecting.
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		shakeHands:       NOPHandshakeFunc,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	print("Listening on ", t.ListenAddr)
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
	peer := NewTCPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("failed to shake hands: %s\n", err)
		return
	}

	msg := &Message{}
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read message: %s\n", err)
			return

		}
		lineReader := bytes.NewReader([]byte(line))

		if err := t.Decoder.Decode(lineReader, msg); err != nil {
			fmt.Printf("failed to decode message: %s\n", err)
			continue
		}
		msg.From = conn.RemoteAddr()
		fmt.Printf("received message: %+v\n", msg)
	}
}
