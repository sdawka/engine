package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// WorkStore is meant to implement a basic store for work.
// This needs a key value interface with some list for discovering
// available games. The lock interface just needs a map with the
// ability to set a TTL. Redis would satisfy this.
type WorkStore interface {
	Lock(ctx context.Context, key string) (string, error)
	Unlock(ctx context.Context, key string) error
	Pop(context.Context) (string, error)

	// As far as writing information, the locking interface requires
	// that the write op for a game tick have a CAS type operation to
	// make sure that a worker hasn't lost a lock in between the time
	// it last checked and the write occuring.
}

// InMemStore returns a new in memory store. Currently hacked with a list of gameID's.
func InMemStore(games ...string) WorkStore {
	return &inmem{
		games: games,
		locks: map[string]*lock{},
	}
}

type lock struct {
	token   string
	expires time.Time
}

type inmem struct {
	games []string
	locks map[string]*lock
	lock  sync.Mutex
}

func (in *inmem) Lock(ctx context.Context, key string) (string, error) {
	in.lock.Lock()
	defer in.lock.Unlock()

	token := ContextGetToken(ctx)
	l, ok := in.locks[key]
	if ok {
		if l.token == token {
			l.expires = time.Now().Add(1 * time.Second)
			return l.token, nil
		}
		if l.expires.Before(time.Now()) {
			delete(in.locks, key)
		} else {
			return "", fmt.Errorf("token expires at %v", l.expires.Unix())
		}
	}
	l = &lock{
		token:   fmt.Sprint(rand.Int63()),
		expires: time.Now().Add(1 * time.Second),
	}
	in.locks[key] = l
	return l.token, nil
}

func (in *inmem) Unlock(ctx context.Context, key string) error {
	in.lock.Lock()
	defer in.lock.Unlock()

	token := ContextGetToken(ctx)
	l, ok := in.locks[key]
	if !ok {
		return nil
	}
	if l.token == token {
		delete(in.locks, key)
	}
	if l.expires.Before(time.Now()) {
		delete(in.locks, key)
	}
	return nil
}

func (in *inmem) Pop(ctx context.Context) (string, error) {
	return in.games[rand.Intn(len(in.games))], nil
}

type contextKey int

const tokenKey contextKey = 1

// ContextWithToken annotates a context with a lock token.
func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// ContextGetToken get's the token from a context.
func ContextGetToken(ctx context.Context) string {
	if s, ok := ctx.Value(tokenKey).(string); ok {
		return s
	}
	return ""
}

// Worker is the worker interface. It wraps a main Perform functions which is
// where all of the game logic should live.
type Worker struct {
	WorkStore         WorkStore
	PollInterval      time.Duration
	HeartbeatInterval time.Duration
	Perform           func(context.Context, string, int) error
}

// Run will run the worker in a loop.
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
