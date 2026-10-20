package actor_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"AxiomLang/stdlib/actor"
)

func TestActorMessagePassing(t *testing.T) {
	system := actor.NewSystem()

	var counter int64

	handler := func(ctx context.Context, msg actor.Message) error {
		if msg.Type == "INCREMENT" {
			atomic.AddInt64(&counter, 1)
		}
		return nil
	}

	_, err := system.Spawn("worker_1", 100, handler)
	if err != nil {
		t.Fatalf("failed to spawn actor: %v", err)
	}

	for i := 0; i < 50; i++ {
		err := system.Send("worker_1", actor.Message{Type: "INCREMENT"})
		if err != nil {
			t.Fatalf("failed to send message: %v", err)
		}
	}

	time.Sleep(50 * time.Millisecond)

	val := atomic.LoadInt64(&counter)
	if val != 50 {
		t.Fatalf("expected counter 50, got %d", val)
	}
}
