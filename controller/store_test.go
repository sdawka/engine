package controller

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	"github.com/stretchr/testify/require"
)

func testStoreLock(t *testing.T, s Store) {
	ctx := context.Background()

	// Lock random key.
	tok, err := s.Lock(ctx, "test", "")
	require.Nil(t, err)
	require.NotEmpty(t, tok)

	// Lock with valid token, no error same token returned.
	tok2, err := s.Lock(ctx, "test", tok)
	require.Nil(t, err)
	require.Equal(t, tok, tok2)

	// Unlock without valid token returns error.
	err = s.Unlock(ctx, "test", "")
	require.NotNil(t, err)

	// Unlock with valid token no error.
	err = s.Unlock(ctx, "test", tok)
	require.Nil(t, err)

	// Unlock where lock doesn't exist returns no error.
	err = s.Unlock(ctx, "missing", "")
	require.Nil(t, err)
}

func testStoreLockExpiry(t *testing.T, s Store) {
	ctx := context.Background()

	// Negative expiry, will always be expired.
	LockExpiry = -10 * time.Second

	// Lock random key.
	tok, err := s.Lock(ctx, "test", "")
	require.Nil(t, err)
	require.NotEmpty(t, tok)

	// Lock (with token) has expired.
	tok2, err := s.Lock(ctx, "test", tok)
	require.Nil(t, err)
	require.Equal(t, tok, tok2)

	// Unlock (no token) has expired.
	err = s.Unlock(ctx, "test", "")
	require.Nil(t, err)

	// Lock (no token) has expired.
	_, err = s.Lock(ctx, "test", "")
	require.Nil(t, err)

	// Unlock (no token) has expired.
	err = s.Unlock(ctx, "test", "")
	require.Nil(t, err)

	// Reset.
	LockExpiry = 1 * time.Second
}

func testStoreGames(t *testing.T, s Store) {
	ctx := context.Background()

	// Create and fetch a game.
	err := s.CreateGame(ctx, &pb.Game{ID: "test", Status: string(rules.GameStatusRunning)}, nil)
	require.Nil(t, err)
	g, err := s.GetGame(ctx, "test")
	require.Nil(t, err)
	require.Equal(t, "test", g.ID)

	// NotFound error thrown.
	_, err = s.GetGame(ctx, "tes11221t")
	require.Equal(t, ErrNotFound, err)

	// Pop game can find it.
	id, err := s.PopGameID(ctx)
	require.Nil(t, err)
	require.Equal(t, "test", id)

	// Lock test key, cannot pop.
	_, err = s.Lock(ctx, "test", "")
	require.Nil(t, err)
	_, err = s.PopGameID(ctx)
	require.NotNil(t, err)
}

func testStoreGameFrames(t *testing.T, s Store) {
	ctx := context.Background()

	// Create and fetch a game.
	err := s.CreateGame(ctx, &pb.Game{ID: "test", Status: string(rules.GameStatusRunning)}, nil)
	require.Nil(t, err)
	g, err := s.GetGame(ctx, "test")
	require.Nil(t, err)
	require.Equal(t, "test", g.ID)

	// Read game frames, too high offset.
	frames, err := s.ListGameFrames(ctx, "test", 10, 100)
	require.Nil(t, err)
	require.Equal(t, 0, len(frames))

	// Read game frames, 0 offset.
	frames, err = s.ListGameFrames(ctx, "test", 10, 0)
	require.Nil(t, err)
	require.Equal(t, 0, len(frames))

	// Push a game frame.
	err = s.PushGameFrame(ctx, "test", &pb.GameFrame{})
	require.Nil(t, err)

	// Read the game frames.
	frames, err = s.ListGameFrames(ctx, "test", 1, 0)
	require.Nil(t, err)
	require.Equal(t, 1, len(frames))

	// Read game frames that don't exist.
	frames, err = s.ListGameFrames(ctx, "test22", 1, 0)
	require.Equal(t, ErrNotFound, err)
	require.Equal(t, 0, len(frames))

	// Read the game frames, too high offset.
	frames, err = s.ListGameFrames(ctx, "test", 10, 100)
	require.Nil(t, err)
	require.Equal(t, 0, len(frames))
}

func testStoreConcurrentWriters(t *testing.T, s Store) {
	ctx := context.Background()

	// Create and fetch a game.
	err := s.CreateGame(ctx, &pb.Game{ID: "test", Status: string(rules.GameStatusRunning)}, nil)
	require.Nil(t, err)

	var ok uint32 // How many got the lock.
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 20; i++ {
		go func() {
			// Lock key, push allowed.
			_, errl := s.Lock(ctx, "test", "")
			// If locked, push should be allowed. If not locked, push not
			// allowed.
			if errl == nil {
				atomic.AddUint32(&ok, 1)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	require.Equal(t, uint32(1), ok)
}

func TestStore_InMem_Lock(t *testing.T)              { testStoreLock(t, InMemStore()) }
func TestStore_InMem_LockExpiry(t *testing.T)        { testStoreLockExpiry(t, InMemStore()) }
func TestStore_InMem_Games(t *testing.T)             { testStoreGames(t, InMemStore()) }
func TestStore_InMem_GameFrames(t *testing.T)        { testStoreGameFrames(t, InMemStore()) }
func TestStore_InMem_ConcurrentWriters(t *testing.T) { testStoreConcurrentWriters(t, InMemStore()) }
