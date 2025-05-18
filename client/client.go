// Package client provides a simple Redis client implementation.
package client

import (
	"bytes"   // for buffer manipulation
	"context" // for context handling
	"io"
	"net" // for network operations

	"github.com/tidwall/resp" // RESP protocol library
)

// Client represents a client connection to a Redis server.
type Client struct {
	addr string // the address of the Redis server (e.g., "localhost:6379")
}

// New creates and initializes a new Client instance.
// address: the network address of the Redis server.
func New(address string) *Client {
	return &Client{
		addr: address,
	}
}

// Set sends a SET command to the Redis server.
// ctx: context for cancellation and timeouts (though not fully utilized here).
// key: the key to set.
// val: the value to set.
// Returns an error if the operation fails.
func (c *Client) Set(ctx context.Context, key string, val string) error {
	// Establish a TCP connection to the server address.
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err // Return any connection error.
	}
	// Use a buffer to build the RESP message.
	buf := &bytes.Buffer{}
	// Create a RESP writer that writes to the buffer.
	wr := resp.NewWriter(buf)
	// Write the SET command as a RESP Array: ["SET", key, val]
	wr.WriteArray([]resp.Value{resp.StringValue("SET"), resp.StringValue(key), resp.StringValue(val)})
	// Write the buffered RESP message to the network connection.
	// _, err = conn.Write(buf.Bytes())
	_, err = io.Copy(conn, buf)
	// Return the error from the write operation (or nil on success).
	return err
}
