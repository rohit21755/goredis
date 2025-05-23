package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/rohit21755/goredis/client"
)

func TestServerWithMultiClinets(t *testing.T) {
	server := NewServer(Config{})
	go func() {

		log.Fatal(server.Start())
	}()
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done() // safer: defer ensures it always gets called

			c, err := client.New("localhost:8000")
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
	wg.Wait()

	if len(server.peers) != 0 {
		t.Fatalf("expected 0 peer")
	}
	time.Sleep(time.Second)
}
