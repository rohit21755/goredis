// Package main implements a TCP server and its components
package main

import (
	"net" // provides basic networking interface
)

// Peer represents a connected client in the TCP server
type Peer struct {
	conn  net.Conn // underlying network connection
	msgCh chan []byte
}

// NewPeer creates and initializes a new Peer instance
// conn: the established TCP connection for this peer
func NewPeer(conn net.Conn, msgCh chan []byte) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

// readLoop continuously reads data from the peer's connection
// It runs in a loop until an error occurs or connection closes
func (p *Peer) readLoop() error {
	buf := make([]byte, 1024) // temporary buffer for reading data
	for {
		// Read incoming data into buffer
		n, err := p.conn.Read(buf)
		if err != nil {
			return err // return any read errors (including EOF)
		}

		// Create a new buffer with exact size of message
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n]) // copy only the bytes that were read
		//- We'd keep the entire buf with 1019 unused bytes
		//- Could contain garbage data from previous reads
		p.msgCh <- msgBuf
		// TODO: Process the received message (msgBuf)
	}
}
