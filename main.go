// Package main implements a TCP server
package main

import (
	// for context handling
	"flag"
	"fmt" // for formatted I/O

	// for logging fatal errors
	"log/slog" // structured logging
	"net"      // networking primitives
	// for timing operations, like sleeps
	// client package for testing
)

// Default TCP server port if none specified
const defaultListenAddr = ":8080"

// Config holds server configuration options
type Config struct {
	ListenerAddr string // TCP address for the server to listen on
}

type Message struct {
	cmd Command
	// conn net.Conn
	peer *Peer
}

// Server represents the TCP server instance
type Server struct {
	Config                   // embedded config struct
	peers     map[*Peer]bool // track active peer connections
	ln        net.Listener   // TCP listener instance
	addPeerCh chan *Peer     // channel to add new peers to the map
	delPeerCh chan *Peer
	quitCh    chan struct{} // channel to signal server shutdown
	msgCh     chan Message  // channel to receive raw messages from peers
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
		delPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}), // initialize quit channel
		msgCh:     make(chan Message),  // initialize message channel
		kv:        NewKV(),             // initialize KV store
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
	slog.Info("go redis Server Running", "listenAddr", s.ListenerAddr)
	// start accepting connections in background

	s.acceptLoop()
	return nil // return nil on successful start
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
func (s *Server) handleMessage(rawMsg Message) error {
	// slog.Info("rawsmg", rawMsg)
	// cmd, err := parseCommand(string(rawMsg.data)) // parse the raw message into a command
	// if err != nil {
	// 	return err // return parsing errors
	// }
	switch v := rawMsg.cmd.(type) {
	case SetCommand:
		// Handle SetCommand by calling the set method
		return s.kv.Set(v.key, v.val)
	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		_, err := rawMsg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}

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
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("Raw message error", "err", err)
			}

		case peer := <-s.addPeerCh: // wait for new peers
			slog.Info("new peer connnected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true // add new peer to the map
			// slog.Info("peer added", "remoteAddr", peer.conn.RemoteAddr())
			// default:
			// 	fmt.Println("foo") // placeholder default case

		case peer := <-s.delPeerCh:
			slog.Info("Peer disconnected")
			delete(s.peers, peer)
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
	newPeer := NewPeer(conn, s.msgCh, s.delPeerCh) // Create a new peer for the connection

	s.addPeerCh <- newPeer // Add the new peer to the server's peer map
	// slog.Info("new connection", "remoteAddr", conn.RemoteAddr())
	// Start the peer's read loop in a goroutine. This will block until the connection is closed or an error occurs.
	if err := newPeer.readLoop(); err != nil {
		slog.Error("Peer read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}

}

func main() {
	listenAddr := flag.String("listenAddr", defaultListenAddr, "listen address for go redis server")
	flag.Parse()
	println(*listenAddr)
	// Start the server in a goroutine so the main function can continue.
	s := NewServer(Config{
		ListenerAddr: *listenAddr,
	})

	s.Start()
	// go func() {

	// 	if err := s.Start(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// // Create and use a client to test the server. This is temporary for testing.
	// client, err := client.New("localhost:8080")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for i := 0; i < 10; i++ {
	// 	// Create a new client instance for each iteration
	// 	fmt.Println("SET =>", fmt.Sprintf("bar_%d", i))
	// 	// Call the Set method on the client
	// 	// Using context.TODO() as context handling is not yet implemented in the client.
	// 	if err := client.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	val, err := client.Get(context.TODO(), fmt.Sprintf("foo_%d", i))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println("GET =>", val)

	// }

	// Keep the main goroutine running for a short duration to allow server to process requests.

	// The program will exit after the sleep finishes, as select{} is commented out.
	// In a real server, you would use select{} or a similar mechanism to keep the main goroutine alive
	// until a shutdown signal is received.
	// select {} // so the program does not exit
}
