package rules

import (
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestUpdateFood(t *testing.T) {
	updated := updateFood(20, 20, &pb.GameTick{
		Food: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
		},
		Snakes: []*pb.Snake{
			{
				Body: []*pb.Point{
					{X: 1, Y: 2},
				},
			},
		},
	}, []*pb.Point{
		{X: 1, Y: 2},
	})
	require.Len(t, updated, 2)
	require.False(t, updated[1].Equal(&pb.Point{X: 1, Y: 2}))
}

func TestGameTickUpdatesTurnCounter(t *testing.T) {
	game := &pb.Game{
		Ticks: []*pb.GameTick{
			&pb.GameTick{Turn: 5},
		},
	}
	gt, err := GameTick(game)
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
		Ticks: []*pb.GameTick{
			&pb.GameTick{
				Turn: 5,
				Snakes: []*pb.Snake{
					snake,
				},
			},
		},
	}
	gt, err := GameTick(game)
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int64(66), snake.Health)
	require.Len(t, snake.Body, 3)
	require.Equal(t, &pb.Point{X: 1, Y: 0}, snake.Body[0])
	require.Equal(t, &pb.Point{X: 1, Y: 1}, snake.Body[1])
	require.Equal(t, &pb.Point{X: 1, Y: 2}, snake.Body[2])
}

var game = &pb.Game{
	Width:  20,
	Height: 20,
	Ticks: []*pb.GameTick{
		&pb.GameTick{
			Turn:   5,
			Snakes: []*pb.Snake{},
			Food: []*pb.Point{
				{X: 1, Y: 0},
			},
		},
	},
}

func TestGameTickSnakeEats(t *testing.T) {
	snake := &pb.Snake{
		Health: 67,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}

	game.Ticks[0].Snakes = []*pb.Snake{snake}

	gt, err := GameTick(game)
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

	game.Ticks[0].Snakes = []*pb.Snake{snake}

	gt, err := GameTick(game)
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int64(87), snake.Health)
	require.Len(t, snake.Body, 3)
	require.Equal(t, &pb.Point{X: 1, Y: 1}, snake.Body[0])
	require.Equal(t, &pb.Point{X: 1, Y: 2}, snake.Body[1])
	require.Equal(t, &pb.Point{X: 1, Y: 3}, snake.Body[2])
}
