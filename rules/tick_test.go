package rules

import (
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestUpdateFood(t *testing.T) {
	updated := updateFood(&pb.Game{
		Width:  20,
		Height: 20,
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
	require.False(t, updated[1].Equals(&pb.Point{X: 1, Y: 2}))
}

func TestGameTickUpdatesTurnCounter(t *testing.T) {
	game := &pb.Game{
		Width:  20,
		Height: 20,
		Turn:   5,
	}
	GameTick(game)
	require.Equal(t, int64(6), game.Turn)
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
		Turn:   5,
		Snakes: []*pb.Snake{
			snake,
		},
	}
	GameTick(game)
	require.Equal(t, int64(66), snake.Health)
	require.Len(t, snake.Body, 2)
	require.True(t, snake.Body[0].Equals(&pb.Point{X: 1, Y: 1}))
	require.True(t, snake.Body[1].Equals(&pb.Point{X: 1, Y: 2}))
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
	game := &pb.Game{
		Width:  20,
		Height: 20,
		Turn:   5,
		Snakes: []*pb.Snake{
			snake,
		},
		Food: []*pb.Point{
			{X: 1, Y: 1},
		},
	}
	GameTick(game)
	require.Equal(t, int64(100), snake.Health)
	require.Len(t, snake.Body, 3)
}
