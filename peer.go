// Package main implements a TCP server and its components
package main

import (
	"fmt"
	"io"
	"log"
	"net" // provides basic networking interface

	"github.com/tidwall/resp"
)

// Peer represents a connected client in the TCP server
type Peer struct {
	conn  net.Conn // underlying network connection
	msgCh chan Message
}

// NewPeer creates and initializes a new Peer instance
// conn: the established TCP connection for this peer
func NewPeer(conn net.Conn, msgCh chan Message) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

// readLoop continuously reads data from the peer's connection
// It runs in a loop until an error occurs or connection closes
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)
	for {
		// Read the next value from the RESP stream
		v, _, err := rd.ReadValue()
		// Check for end of file (end of input)
		if err == io.EOF {
			break // exit the loop if input is consumed
		}
		// Handle other potential read errors
		if err != nil {
			log.Fatal(err) // log fatal errors and exit (consider more graceful error handling)
		}
		// Check if the parsed value is a RESP Array (typical for commands)
		if v.Type() == resp.Array {
			// Iterate through the elements of the array
			for _, value := range v.Array() {
				// Check the command name (first element of the array)
				switch value.String() {
				case CommandGET:
					// Validate the number of arguments for the SET command
					if len(v.Array()) != 2 {
						return fmt.Errorf("Invalid command") // return error for incorrect argument count
					}
					// Create and return a SetCommand struct
					cmd := GetCommand{

						key: v.Array()[1].Bytes(), // Extract the key
						// Extract the value
					}

					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}
					// return cmd, nil
				case CommandSET:
					// Validate the number of arguments for the SET command
					if len(v.Array()) != 3 {
						return fmt.Errorf("Invalid command") // return error for incorrect argument count
					}
					// Create and return a SetCommand struct
					cmd := SetCommand{

						key: v.Array()[1].Bytes(), // Extract the key
						val: v.Array()[2].Bytes(), // Extract the value
					}

					p.msgCh <- Message{
						cmd:  cmd,
						peer: p,
					}
					// return cmd, nil // Return the parsed command and no error
				default:
					// Handle unknown commands (can add more cases here)
				}
			}
		}
		// Return an error for invalid or unknown command format
		// return nil, fmt.Errorf("invalid or unknown command")
	}
	// Return an error if no command was parsed from the input
	// return nil, fmt.Errorf("invalid or unknown command")
	return nil

}
