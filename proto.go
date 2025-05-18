// Package main implements Redis protocol parsing and command handling
package main

import (
	"bytes" // for byte buffer manipulation
	"fmt"   // for formatted I/O
	"io"    // for EOF handling
	"log"   // for logging fatal errors

	"github.com/tidwall/resp" // RESP protocol parsing library
)

// Constants for command names
const (
	CommandSET = "SET"
)

// Command is an interface for all command types
type Command interface {
	// Methods specific to commands can be added here later
}

// SetCommand represents the Redis SET command
type SetCommand struct {
	key, val []byte // key and value to be set
}

// parseCommand parses a raw byte slice representing a Redis command into a Command interface
// raw: the raw byte slice received from a client
func parseCommand(raw string) (Command, error) {
	// Create a new RESP reader from the raw input
	rd := resp.NewReader(bytes.NewBufferString(raw))
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
				case CommandSET:
					// Validate the number of arguments for the SET command
					if len(v.Array()) != 3 {
						return nil, fmt.Errorf("Invalid command") // return error for incorrect argument count
					}
					// Create and return a SetCommand struct
					cmd := SetCommand{

						key: v.Array()[1].Bytes(), // Extract the key
						val: v.Array()[2].Bytes(), // Extract the value
					}
					return cmd, nil // Return the parsed command and no error
				default:
					// Handle unknown commands (can add more cases here)
				}
			}
		}
		// Return an error for invalid or unknown command format
		return nil, fmt.Errorf("invalid or unknown command")
	}
	// Return an error if no command was parsed from the input
	return nil, fmt.Errorf("invalid or unknown command")

}
