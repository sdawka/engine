package worker

import (
	"context"
	"fmt"
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
}

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
			return "", fmt.Errorf("token expires at %v", l.expires)
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
