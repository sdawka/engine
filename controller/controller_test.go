package controller

import (
	"context"
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

var client pb.ControllerClient

func init() {
	ctrl := New(InMemStore())
	go func() {
		if err := ctrl.Serve(":0"); err != nil {
			panic(err)
		}
	}()
	var err error
	client, err = pb.Dial(ctrl.DialAddress())
	if err != nil {
		panic(err)
	}
}

func TestController_Lock(t *testing.T) {
	ctx := context.Background()
	ctrl := client

	// Lock key (game doesn't need to exist).
	tok, err := ctrl.Lock(ctx, &pb.LockRequest{ID: "test"})
	require.Nil(t, err)

	// Lock again (without token).
	_, err = ctrl.Lock(ctx, &pb.LockRequest{ID: "test"})
	require.NotNil(t, err)

	// Lock again (with token).
	ctx = pb.ContextWithLockToken(ctx, tok.Token)
	_, err = ctrl.Lock(ctx, &pb.LockRequest{ID: "test"})
	require.Nil(t, err)

	// Unlock (with token).
	_, err = ctrl.Unlock(ctx, &pb.UnlockRequest{ID: "test"})
	require.Nil(t, err)
}

func TestController_Games(t *testing.T) {
	ctx := context.Background()
	ctrl := client

	// Start test game.
	resp, _ := ctrl.Create(context.Background(), &pb.CreateRequest{})
	_, err := ctrl.Start(context.Background(), &pb.StartRequest{
		ID: resp.ID,
	})
	require.Nil(t, err)

	// Should pop above game.
	g, err := ctrl.Pop(ctx, &pb.PopRequest{})
	require.Nil(t, err)
	require.Equal(t, resp.ID, g.ID)

	// Should get above game.
	game, err := ctrl.Status(ctx, &pb.StatusRequest{
		ID: g.ID,
	})
	require.Nil(t, err)
	require.Equal(t, "running", game.Game.Status)
	require.Equal(t, int64(1000), game.Game.SnakeTimeout)
	require.Equal(t, "multi-player", game.Game.Mode)
	require.NotNil(t, game.LastTick)
	require.Equal(t, int64(0), game.LastTick.Turn)
}

func TestController_PushTicks(t *testing.T) {
	ctx := context.Background()
	ctrl := client

	// Start test game.
	g, _ := ctrl.Create(context.Background(), &pb.CreateRequest{})
	_, err := ctrl.Start(context.Background(), &pb.StartRequest{
		ID: g.ID,
	})
	require.Nil(t, err)

	// Lock the game.
	tok, err := ctrl.Lock(ctx, &pb.LockRequest{ID: g.ID})
	require.Nil(t, err)
	ctx = pb.ContextWithLockToken(ctx, tok.Token)

	// Should get above game.
	_, err = ctrl.AddGameTick(ctx, &pb.AddGameTickRequest{
		ID:       g.ID,
		GameTick: &pb.GameTick{},
	})
	require.Nil(t, err)

	// List a set of game ticks.
	ticks, err := ctrl.ListGameTicks(ctx, &pb.ListGameTicksRequest{ID: g.ID, Limit: 10})
	require.Nil(t, err)
	require.Equal(t, 2, len(ticks.Ticks))
	require.Equal(t, 2, int(ticks.Count))
}

func TestController_EndGame(t *testing.T) {
	ctx := context.Background()
	ctrl := client

	// Start test game.
	g, _ := ctrl.Create(ctx, &pb.CreateRequest{})
	_, err := ctrl.Start(ctx, &pb.StartRequest{ID: g.ID})
	require.Nil(t, err)

	// End test game.
	_, err = ctrl.EndGame(ctx, &pb.EndGameRequest{ID: g.ID})
	require.Nil(t, err)
}
