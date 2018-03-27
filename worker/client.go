package worker

import (
	"context"
	"fmt"
	"log"
	"time"
)

type contextKey int

const tokenKey contextKey = 1

func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func ContextGetToken(ctx context.Context) string {
	if s, ok := ctx.Value(tokenKey).(string); ok {
		return s
	}
	return ""
}

type Worker struct {
	WorkStore         WorkStore
	PollInterval      time.Duration
	HeartbeatInterval time.Duration
	Perform           func(context.Context, string, int) error
}

func (w *Worker) Run(workerID int) {
	for {
		// We are now holding the lock.
		if err := w.run(context.Background(), workerID); err != nil {
			log.Printf("run failed: %v", err)
			time.Sleep(w.PollInterval)
		}
	}
}

func (w *Worker) run(ctx context.Context, workerID int) error {
	// Pop an item of work.
	id, err := w.WorkStore.Pop(ctx)
	if err != nil {
		return fmt.Errorf("no work found: %v", err)
	}
	// Attempt to get the lock initially.
	token, err := w.WorkStore.Lock(ctx, id)
	if err != nil {
		return fmt.Errorf("not able to acquire lock: %v", err)
	}

	log.Printf("[%d] acquired lock %s token=%s", workerID, id, token)

	// Get a context with the lock token.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = ContextWithToken(ctx, token)
	defer w.WorkStore.Unlock(ctx, id)

	// Hold the lock, heartbeating every HeartbeatInterval.
	go func() {
		t := time.NewTicker(w.HeartbeatInterval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_, err := w.WorkStore.Lock(ctx, id)
				if err != nil {
					log.Printf("[%d] lock expired during heartbeat %v", workerID, err)
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Perform the actual work, this should respect context and Done() rules.
	// Perform should be able to write to storage using the context and have
	// a valid lock for the key.
	return w.Perform(ctx, id, workerID)
}
