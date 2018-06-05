package rules

import (
	"errors"
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestUpdateFood(t *testing.T) {
	updated, err := updateFood(20, 20, &pb.GameFrame{
		Food: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
		},
		Snakes: []*pb.Snake{
			{
				Body: []*pb.Point{
					{X: 1, Y: 2},
					{X: 2, Y: 2},
					{X: 3, Y: 2},
				},
			},
		},
	}, []*pb.Point{
		{X: 1, Y: 2},
	})
	require.NoError(t, err)
	require.Len(t, updated, 2)
	require.True(t, updated[0].Equal(&pb.Point{X: 1, Y: 1}))
	require.False(t, updated[1].Equal(&pb.Point{X: 1, Y: 2}))
	require.False(t, updated[1].Equal(&pb.Point{X: 2, Y: 2}))
	require.False(t, updated[1].Equal(&pb.Point{X: 3, Y: 2}))
	require.False(t, updated[1].Equal(&pb.Point{X: 1, Y: 1}))
}

func TestGameTickUpdatesTurnCounter(t *testing.T) {
	gt, err := GameTick(commonGame, &pb.GameFrame{Turn: 5})
	require.NoError(t, err)
	require.Equal(t, int64(6), gt.Turn)
}

func TestGameTickUpdatesSnake(t *testing.T) {
	snake := &pb.Snake{
		Health: 67,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}
	game := &pb.Game{
		Width:  20,
		Height: 20,
	}
	gt, err := GameTick(game, &pb.GameFrame{
		Turn: 5,
		Snakes: []*pb.Snake{
			snake,
		},
	})
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int64(66), snake.Health)
	require.Len(t, snake.Body, 3)
	require.Equal(t, &pb.Point{X: 1, Y: 0}, snake.Body[0])
	require.Equal(t, &pb.Point{X: 1, Y: 1}, snake.Body[1])
	require.Equal(t, &pb.Point{X: 1, Y: 2}, snake.Body[2])
}

var commonGame = &pb.Game{
	Width:  20,
	Height: 20,
}
var lastFrame = &pb.GameFrame{
	Turn:   5,
	Snakes: []*pb.Snake{},
	Food: []*pb.Point{
		{X: 1, Y: 0},
	},
}

func TestGameFrameSnakeEats(t *testing.T) {
	snake := &pb.Snake{
		Health: 67,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}

	lastFrame.Snakes = []*pb.Snake{snake}

	gt, err := GameTick(commonGame, lastFrame)
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int64(100), snake.Health)
	require.Len(t, snake.Body, 4)
}

func TestGameTickDeadSnakeDoNotUpdate(t *testing.T) {
	snake := &pb.Snake{
		Health: 87,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
		Death: &pb.Death{
			Turn:  4,
			Cause: DeathCauseSnakeCollision,
		},
	}

	lastFrame.Snakes = []*pb.Snake{snake}

	gt, err := GameTick(commonGame, lastFrame)
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int64(87), snake.Health)
	require.Len(t, snake.Body, 3)
	require.Equal(t, &pb.Point{X: 1, Y: 1}, snake.Body[0])
	require.Equal(t, &pb.Point{X: 1, Y: 2}, snake.Body[1])
	require.Equal(t, &pb.Point{X: 1, Y: 3}, snake.Body[2])
}

func TestGameTickUpdatesDeath(t *testing.T) {
	snake := &pb.Snake{
		Health: 0,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}

	lastFrame.Snakes = []*pb.Snake{snake}

	gt, err := GameTick(commonGame, lastFrame)
	require.NoError(t, err)
	require.NotNil(t, gt.Snakes[0].Death)
}

func TestUpdateSnakes(t *testing.T) {
	snake := &pb.Snake{
		Body: []*pb.Point{
			{X: 1, Y: 1},
		},
	}
	moves := []*SnakeUpdate{
		&SnakeUpdate{
			Snake: snake,
			Err:   errors.New("some error"),
		},
	}
	updateSnakes(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	}, moves)
	require.Equal(t, &pb.Point{X: 1, Y: 0}, snake.Head(), "snake did not move up")

	moves = []*SnakeUpdate{
		&SnakeUpdate{
			Snake: snake,
			Move:  "left",
		},
	}
	updateSnakes(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	}, moves)
	require.Equal(t, &pb.Point{X: 0, Y: 0}, snake.Head(), "snake did not move left")
}
