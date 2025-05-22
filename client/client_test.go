package client

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c, err := New("localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		// Create a new client instance for each iteration
		fmt.Println("SET =>", fmt.Sprintf("bar_%d", i))
		// Call the Set method on the client
		// Using context.TODO() as context handling is not yet implemented in the client.
		if err := c.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}

		val, err := c.Get(context.TODO(), fmt.Sprintf("foo_%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("GET =>", val)

	}
}
