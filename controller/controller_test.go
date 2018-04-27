package controller

import (
	"context"
	"sync"
	"sync/atomic"
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

func TestController_GameCRUD(t *testing.T) {
	ctx := context.Background()

	var gameID string
	var token string

	t.Run("StartGame_NoGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("Create", func(t *testing.T) {
		resp, err := client.Create(ctx, &pb.CreateRequest{})
		require.Nil(t, err)
		gameID = resp.ID
	})

	t.Run("StartGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.Nil(t, err)
	})

	t.Run("PopGame", func(t *testing.T) {
		resp, err := client.Pop(ctx, &pb.PopRequest{})
		require.Nil(t, err)
		require.Equal(t, gameID, resp.ID)
		token = resp.Token
	})

	t.Run("PopGame_NoGames", func(t *testing.T) {
		_, err := client.Pop(ctx, &pb.PopRequest{})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("GetGame", func(t *testing.T) {
		game, err := client.Status(ctx, &pb.StatusRequest{ID: gameID})
		require.Nil(t, err)
		require.Equal(t, "running", game.Game.Status)
		require.Equal(t, int64(1000), game.Game.SnakeTimeout)
		require.Equal(t, "multi-player", game.Game.Mode)
		require.NotNil(t, game.LastTick)
		require.Equal(t, int64(0), game.LastTick.Turn)
	})

	t.Run("GetGame_NoGames", func(t *testing.T) {
		_, err := client.Status(ctx, &pb.StatusRequest{
			ID: "non-existant",
		})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("EndGame_NoToken", func(t *testing.T) {
		_, err := client.EndGame(ctx, &pb.EndGameRequest{ID: gameID})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = ResourceExhausted desc = controller: game is locked",
			err.Error(),
		)
	})

	t.Run("EndGame", func(t *testing.T) {
		_, err := client.EndGame(
			pb.ContextWithLockToken(ctx, token), &pb.EndGameRequest{ID: gameID})
		require.Nil(t, err)
	})
}

func TestController_Ticks(t *testing.T) {
	ctx := context.Background()

	var gameID string
	var token string

	t.Run("Create", func(t *testing.T) {
		resp, err := client.Create(ctx, &pb.CreateRequest{})
		require.Nil(t, err)
		gameID = resp.ID
	})

	t.Run("StartGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.Nil(t, err)
	})

	t.Run("PopGame", func(t *testing.T) {
		resp, err := client.Pop(ctx, &pb.PopRequest{})
		require.Nil(t, err)
		require.Equal(t, gameID, resp.ID)
		token = resp.Token
	})

	t.Run("AddGameTick_NoGame", func(t *testing.T) {
		_, err := client.AddGameFrame(ctx, &pb.AddGameFrameRequest{
			ID:        "foo",
			GameFrame: &pb.GameFrame{},
		})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("AddGameTick_NoLock", func(t *testing.T) {
		_, err := client.AddGameFrame(ctx, &pb.AddGameFrameRequest{
			ID:        gameID,
			GameFrame: &pb.GameFrame{},
		})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = ResourceExhausted desc = controller: game is locked",
			err.Error(),
		)
	})

	t.Run("AddGameTick_Nil", func(t *testing.T) {
		_, err := client.AddGameFrame(
			pb.ContextWithLockToken(ctx, token),
			&pb.AddGameFrameRequest{
				ID:        gameID,
				GameFrame: nil,
			})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = InvalidArgument desc = controller: game tick must not be nil",
			err.Error(),
		)
	})

	t.Run("AddGameTicks", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			resp, err := client.AddGameFrame(
				pb.ContextWithLockToken(ctx, token), &pb.AddGameFrameRequest{
					ID:        gameID,
					GameFrame: &pb.GameFrame{},
				})
			require.Nil(t, err)
			require.Equal(t, gameID, resp.Game.ID)
		}
	})

	t.Run("ListGameTicks_NoGame", func(t *testing.T) {
		_, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{ID: "foo"})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("ListGameTicks_BadOffset", func(t *testing.T) {
		ticks, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{
			ID: gameID, Offset: 1000000})
		require.Nil(t, err)
		require.Empty(t, ticks.Ticks)
	})

	t.Run("ListGameTicks", func(t *testing.T) {
		ticks, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{ID: gameID})
		require.Nil(t, err)
		// Returns max 50.
		require.Equal(t, 50, len(ticks.Ticks))
		require.Equal(t, 50, int(ticks.Count))
	})

	t.Run("ListGameTicks_Min50", func(t *testing.T) {
		ticks, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{
			ID:    gameID,
			Limit: 1000,
		})
		require.Nil(t, err)
		// Returns max 50.
		require.Equal(t, 50, len(ticks.Ticks))
		require.Equal(t, 50, int(ticks.Count))
	})
}

func TestController_PopConcurrent(t *testing.T) {
	ctx := context.Background()
	ctrl := client

	g, _ := ctrl.Create(ctx, &pb.CreateRequest{})
	_, err := ctrl.Start(ctx, &pb.StartRequest{ID: g.ID})
	require.Nil(t, err)

	var ok uint32 // How many got the lock.
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			_, errp := ctrl.Pop(ctx, &pb.PopRequest{})
			if errp == nil {
				atomic.AddUint32(&ok, 1)
			} else {
				require.Equal(t,
					"rpc error: code = NotFound desc = controller: game not found",
					errp.Error(),
				)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	require.Equal(t, uint32(1), ok)
}
