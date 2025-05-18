// Package main implements a TCP server
package main

import (
	"context"  // for context handling
	"fmt"      // for formatted I/O
	"log"      // for logging fatal errors
	"log/slog" // structured logging
	"net"      // networking primitives
	"time"     // for timing operations, like sleeps

	"github.com/rohit21755/goredis/client" // client package for testing
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
	quitCh    chan struct{}  // channel to signal server shutdown
	msgCh     chan []byte    // channel to receive raw messages from peers
	kv        *KV
}

// NewServer creates and initializes a new TCP server instance
// cfg: server configuration
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
		quitCh:    make(chan struct{}),  // initialize quit channel
		msgCh:     make(chan []byte),    // initialize message channel
		kv:        NewKV(),              // initialize KV store
	}
}

// Start initializes the TCP listener and begins accepting connections
// Returns an error if the listener fails to start.
func (s *Server) Start() error {
	// Create TCP listener
	ln, err := net.Listen("tcp", s.ListenerAddr)
	if err != nil {
		return err // return any listener creation errors
	}
	s.ln = ln   // store listener in server struct
	go s.Loop() // start peer management loop in background
	slog.Info("Server Running", "listenAddr", s.ListenerAddr)
	go s.acceptLoop() // start accepting connections in background
	return nil        // return nil on successful start
}

// set handles the SET command
// k: the key to set
// v: the value to set
// Returns an error if the operation fails.
// func (s *Server) set(k, v string) error {
// 	// TODO: Implement logic to set key-value pair in storage
// 	slog.Info("SET command received", "key", k, "value", v)
// 	return nil // Placeholder return
// }

// handleRawMessage parses and handles raw messages from peers
// rawMsg: the raw byte slice received from a peer
// Returns an error if parsing or handling fails.
func (s *Server) handleRawMessage(rawMsg []byte) error {
	// slog.Info("rawsmg", rawMsg)
	cmd, err := parseCommand(string(rawMsg)) // parse the raw message into a command
	if err != nil {
		return err // return parsing errors
	}
	switch v := cmd.(type) {
	case SetCommand:
		// Handle SetCommand by calling the set method
		return s.kv.Set(v.key, v.val)
	// TODO: Add cases for other commands (GET, DEL, etc.)
	default:
		// Log unknown commands
		slog.Info("Unknown command received")
	}

	return nil // return nil for unhandled commands or success
}

// Loop handles server-wide operations like peer management and message processing
func (s *Server) Loop() {
	for {
		select {
		case <-s.quitCh:
			return // Exit loop when quit signal is received
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMessage(rawMsg); err != nil {
				slog.Error("Raw message error", "err", err)
			}

		case peer := <-s.addPeerCh: // wait for new peers
			s.peers[peer] = true // add new peer to the map
			slog.Info("peer added", "remoteAddr", peer.conn.RemoteAddr())
			// default:
			// 	fmt.Println("foo") // placeholder default case
		}
	}
}

// acceptLoop continuously accepts new TCP connections
// It runs in a goroutine and never returns unless the listener is closed.
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept() // accept new connection
		if err != nil {
			slog.Error("Accept error", "err", err)
			continue // continue accepting connections even if one fails
		}
		go s.handleConn(conn) // handle each connection in a goroutine
	}
}

// handleConn processes individual TCP connections
// conn: the established connection for the peer
func (s *Server) handleConn(conn net.Conn) {
	// Create a new peer for the connection and start its read loop.
	newPeer := NewPeer(conn, s.msgCh) // Create a new peer for the connection

	s.addPeerCh <- newPeer // Add the new peer to the server's peer map
	slog.Info("new connection", "remoteAddr", conn.RemoteAddr())
	// Start the peer's read loop in a goroutine. This will block until the connection is closed or an error occurs.
	if err := newPeer.readLoop(); err != nil {
		slog.Error("Peer read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}

}

func main() {
	fmt.Println("Hello, World!") // placeholder main function

	// Start the server in a goroutine so the main function can continue.
	s := NewServer(Config{})
	go func() {

		if err := s.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// Create and use a client to test the server. This is temporary for testing.
	client := client.New("localhost:8080")
	for i := 0; i < 10; i++ {
		// Create a new client instance for each iteration

		// Call the Set method on the client
		// Using context.TODO() as context handling is not yet implemented in the client.
		if err := client.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(s.kv.data)
	// Keep the main goroutine running for a short duration to allow server to process requests.
	time.Sleep(time.Second)

	// The program will exit after the sleep finishes, as select{} is commented out.
	// In a real server, you would use select{} or a similar mechanism to keep the main goroutine alive
	// until a shutdown signal is received.
	// select {} // so the program does not exit
}
