package controller

import (
	"context"
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
	err := s.PutGame(ctx, &pb.Game{ID: "test", Status: rules.GameStatusRunning})
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

func TestStore_InMem_Lock(t *testing.T)       { testStoreLock(t, InMemStore()) }
func TestStore_InMem_LockExpiry(t *testing.T) { testStoreLockExpiry(t, InMemStore()) }
func TestStore_InMem_Games(t *testing.T)      { testStoreGames(t, InMemStore()) }
