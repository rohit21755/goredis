// Package main implements a TCP server
package main

import (
	"fmt"
	"log"
	"log/slog" // structured logging
	"net"      // networking primitives
)

// Default TCP server port if none specified
const defaultListenAddr = ":8080"

// Config holds server configuration options
type Config struct {
	ListenerAddr string // TCP address for the server to listen on
}

// Server represents the TCP server instance
type Server struct {
	Config                   // embedded config struct
	peers     map[*Peer]bool // track active peer connections
	ln        net.Listener   // TCP listener instance
	addPeerCh chan *Peer     // channel to add new peers to the map
	quitCh    chan struct{}
}

// NewServer creates and initializes a new TCP server instance
func NewServer(cfg Config) *Server {
	// Use default address if none provided
	if len(cfg.ListenerAddr) == 0 {
		cfg.ListenerAddr = defaultListenAddr
	}
	// Initialize and return new server with default values
	return &Server{
		Config:    cfg,                  // store the configuration
		peers:     make(map[*Peer]bool), // initialize empty peers map
		addPeerCh: make(chan *Peer),     // create channel for peer management
		quitCh:    make(chan struct{}),
	}
}

// Start initializes the TCP listener and begins accepting connections
func (s *Server) Start() error {
	// Create TCP listener
	ln, err := net.Listen("tcp", s.ListenerAddr)
	if err != nil {
		return err // return any listener creation errors
	}
	s.ln = ln   // store listener in server struct
	go s.Loop() // start peer management loop in background
	slog.Info("Server Running")
	go s.acceptLoop() // start accepting connections in background
	return nil        // return nil on successful start
}

// Loop handles peer management operations
func (s *Server) Loop() {
	for {
		select {
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh: // wait for new peers
			s.peers[peer] = true // add new peer to the map
			// default:
			// 	fmt.Println("foo") // placeholder default case
		}
	}
}

// acceptLoop continuously accepts new TCP connections
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept() // accept new connection
		if err != nil {
			slog.Error("Accept error", "err", err) // log any accept errors
			continue                               // continue accepting connections
		}
		go s.handleConn(conn) // handle each connection in a goroutine
	}
}

// handleConn processes individual TCP connections
func (s *Server) handleConn(conn net.Conn) {
	// TODO: Implement connection handling logic
	newPeer := NewPeer(conn)

	s.addPeerCh <- newPeer
	slog.Info("new connection")
	if err := newPeer.readLoop(); err != nil {
		slog.Error("Peer read error")
	}

}

func main() {
	fmt.Println("Hello, World!") // placeholder main function
	s := NewServer(Config{})
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	// Keep the main goroutine running
	select {}
}
