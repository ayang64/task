package task

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {

	mgr, err := NewManager()

	if err != nil {
		t.Fatal(err)
	}

	words := []string{"beefy", "bananna", "cow", "horse", "sunny", "patty", "grouchy", "hamburger", "flea", "camel", "funky"}

	name := func() string {
		n := []string{}
		for i, max := 0, 2; i < max; i++ {
			n = append(n, words[rand.Intn(len(words))])
		}

		return strings.Join(n, " ")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)

	go func() {
		for i := 0; i < rand.Intn(1<<16); i++ {
			go func() {
				errCh <- mgr.Run(ctx, name(), func(ctx context.Context) error {
					time.Sleep(time.Duration(rand.Intn(50000)) * time.Millisecond)
					return nil
				})
			}()
		}
	}()

	for {
		log.Printf("mgr.PS() = %+v", mgr.PS())
		time.Sleep(1 * time.Second)
	}

}
