package subpub_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"sub-pub/subpub" // адаптируй под свой путь
)

// Проверяет, что подписчик получает опубликованное сообщение.
func TestPublishAndSubscribe(t *testing.T) {
	pubsub := subpub.NewSubPub()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		received []interface{}
	)

	wg.Add(1)
	sub, err := pubsub.Subscribe("topic1", func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg)
		wg.Done()
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	err = pubsub.Publish("topic1", "hello")
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	wg.Wait()
	if len(received) != 1 || received[0] != "hello" {
		t.Errorf("unexpected received messages: %v", received)
	}

	sub.Unsubscribe()
}

// Проверяет, что все подписчики получают сообщение.
func TestMultipleSubscribers(t *testing.T) {
	pubsub := subpub.NewSubPub()
	count := 10
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		_, err := pubsub.Subscribe("topic", func(msg interface{}) {
			if msg != "ping" {
				t.Errorf("expected 'ping', got %v", msg)
			}
			wg.Done()
		})
		if err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
	}

	err := pubsub.Publish("topic", "ping")
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	wg.Wait()
}

// Проверяет, что после отписки сообщения не приходят.
func TestUnsubscribe(t *testing.T) {
	pubsub := subpub.NewSubPub()
	called := false

	sub, err := pubsub.Subscribe("test", func(msg interface{}) {
		called = true
	})
	if err != nil {
		t.Fatal(err)
	}

	err = sub.Unsubscribe()
	if err != nil {
		t.Fatal(err)
	}

	err = pubsub.Publish("test", "data")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	if called {
		t.Error("handler was called after unsubscribe")
	}
}

// Проверяет корректное завершение всех подписок при Close().
func TestCloseWithContext(t *testing.T) {
	pubsub := subpub.NewSubPub()

	_, err := pubsub.Subscribe("close-test", func(msg interface{}) {})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := pubsub.Close(ctx); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestPublishAfterClose(t *testing.T) {
	pubsub := subpub.NewSubPub()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = pubsub.Close(ctx)

	err := pubsub.Publish("key", "data")
	if err == nil {
		t.Error("expected error publishing after Close, got nil")
	}
}

func TestSubscribeAfterClose(t *testing.T) {
	pubsub := subpub.NewSubPub()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = pubsub.Close(ctx)

	_, err := pubsub.Subscribe("key", func(msg interface{}) {})
	if err == nil {
		t.Error("expected error subscribing after Close, got nil")
	}
}
