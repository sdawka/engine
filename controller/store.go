package controller

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/pkg/controller/pb"
)

var (
	// LockExpiry is the time after which a lock will expire.
	LockExpiry = 1 * time.Second
	// ErrNotFound is thrown when a game is not found.
	ErrNotFound = errors.New("controller: game not found")
	// ErrIsLocked is returned when a game is locked.
	ErrIsLocked = errors.New("controller: game is locked")
)

// Store is the interface to the backend store.
type Store interface {
	Lock(ctx context.Context, key, token string) (string, error)
	Unlock(ctx context.Context, key, token string) error
	PopGameID(context.Context) (string, error)
	PutGame(context.Context, *pb.Game) error
	GetGame(context.Context, string) (*pb.Game, error)
}

// InMemStore returns an in memory implementation of the Store interface.
func InMemStore() Store {
	return &inmem{
		games: map[string]*pb.Game{},
		locks: map[string]*lock{},
	}
}

type lock struct {
	token   string
	expires time.Time
}

type inmem struct {
	games map[string]*pb.Game
	locks map[string]*lock
	lock  sync.Mutex
}

func (in *inmem) Lock(ctx context.Context, key, token string) (string, error) {
	in.lock.Lock()
	defer in.lock.Unlock()

	l, ok := in.locks[key]
	if ok {
		if l.token == token {
			l.expires = time.Now().Add(LockExpiry)
			return l.token, nil
		}
		if l.expires.Before(time.Now()) {
			delete(in.locks, key)
		} else {
			return "", ErrIsLocked
		}
	}
	l = &lock{
		token:   fmt.Sprint(rand.Int63()),
		expires: time.Now().Add(LockExpiry),
	}
	in.locks[key] = l
	return l.token, nil
}

func (in *inmem) isLocked(key string) bool {
	l, ok := in.locks[key]
	return ok && l.expires.After(time.Now())
}

func (in *inmem) Unlock(ctx context.Context, key, token string) error {
	in.lock.Lock()
	defer in.lock.Unlock()

	l, ok := in.locks[key]
	if !ok {
		return nil
	}
	if l.token == token {
		delete(in.locks, key)
		return nil
	}
	t := time.Now()
	if l.expires.Before(t) {
		delete(in.locks, key)
		return nil
	}
	return ErrIsLocked
}

func (in *inmem) PopGameID(ctx context.Context) (string, error) {
	in.lock.Lock()
	defer in.lock.Unlock()

	for id := range in.games {
		if !in.isLocked(id) {
			return id, nil
		}
	}
	return "", ErrNotFound
}

func (in *inmem) PutGame(ctx context.Context, g *pb.Game) error {
	in.lock.Lock()
	defer in.lock.Unlock()

	in.games[g.ID] = g
	return nil
}

func (in *inmem) GetGame(ctx context.Context, id string) (*pb.Game, error) {
	in.lock.Lock()
	defer in.lock.Unlock()

	if g, ok := in.games[id]; ok {
		return g, nil
	}
	return nil, ErrNotFound
}
