package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewClient1(t *testing.T) {
	c, err := New("localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	time.Sleep(time.Second)

	// Create a new client instance for each iteration
	fmt.Println("SET =>", "foo")
	// Call the Set method on the client
	// Using context.TODO() as context handling is not yet implemented in the client.
	if err := c.Set(context.TODO(), "foo", 1); err != nil {
		log.Fatal(err)
	}

	val, err := c.Get(context.TODO(), "foo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("GET =>", val)

}

func TestNewClients(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done() // safer: defer ensures it always gets called

			c, err := New("localhost:8000")
			if err != nil {
				t.Fatal(err) // better than log.Fatal in test
			}
			defer c.Close()
			key := fmt.Sprintf("foo_client_%d", i)
			value := fmt.Sprintf("bar_client_%d", i)
			if err := c.Set(context.TODO(), key, value); err != nil {
				t.Fatal(err)
			}

			val, err := c.Get(context.TODO(), key)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("Client %d got this val back => %s\n", i, val)
		}(i)
	}

	wg.Wait() // wait for all goroutines to finish

}
