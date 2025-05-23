// Package client provides a simple Redis client implementation.
package client

import (
	"bytes"   // for buffer manipulation
	"context" // for context handling
	"net"     // for network operations
	"time"

	"github.com/tidwall/resp" // RESP protocol library
)

// Client represents a client connection to a Redis server.
type Client struct {
	addr string // the address of the Redis server (e.g., "localhost:6379")
	conn net.Conn
}

// New creates and initializes a new Client instance.
// address: the network address of the Redis server.
func New(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err // Return any connection error.
	}
	return &Client{
		addr: address,
		conn: conn,
	}, nil
}

// Set sends a SET command to the Redis server.
// ctx: context for cancellation and timeouts (though not fully utilized here).
// key: the key to set.
// val: the value to set.
// Returns an error if the operation fails.
func (c *Client) Set(ctx context.Context, key string, val string) error {
	// Establish a TCP connection to the server address.

	time.Sleep(time.Second)
	// Use a buffer to build the RESP message.
	buf := &bytes.Buffer{}
	// Create a RESP writer that writes to the buffer.
	wr := resp.NewWriter(buf)
	// Write the SET command as a RESP Array: ["SET", key, val]
	wr.WriteArray([]resp.Value{resp.StringValue("SET"), resp.StringValue(key), resp.StringValue(val)})
	// Write the buffered RESP message to the network connection.
	// _, err = conn.Write(buf.Bytes())
	_, err := c.conn.Write(buf.Bytes())
	// Return the error from the write operation (or nil on success).
	return err
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {

	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	wr.WriteArray([]resp.Value{resp.StringValue("GET"), resp.StringValue(key)})
	_, err := c.conn.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	b := make([]byte, 1024)
	n, err := c.conn.Read(b)

	return string(b[:n]), err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
