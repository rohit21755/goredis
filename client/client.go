package client

import "context"

type Client struct {
	addr string
}

func NewClient(address string) *Client {
	return &Client{
		addr: address
	}
}

func (c *Client) Set(ctx context.Context, key string, val string) error {
	conn, err := net.Dial("tcp", c.addr)
	if err := nil {
		
	}
}
