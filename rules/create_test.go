package rules

import (
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestCreateInitialGame(t *testing.T) {
	g, err := CreateInitialGame(&pb.CreateRequest{})
	require.NoError(t, err)
	require.NotNil(t, g.ID)
}

func TestCreateInitialGame_DuplicateSnakeIDs(t *testing.T) {
	_, err := CreateInitialGame(&pb.CreateRequest{
		Width:  20,
		Height: 20,
		Snakes: []*pb.SnakeOptions{
			{ID: "snake_123"},
			{ID: "snake_123"},
		},
	})
	require.Error(t, err)
}

func TestCreateInitialGame_GeneratedSnakeID(t *testing.T) {
	g, err := CreateInitialGame(&pb.CreateRequest{
		Width:  20,
		Height: 20,
		Snakes: []*pb.SnakeOptions{
			{ID: ""},
		},
	})
	require.NoError(t, err)
	require.Len(t, g.Ticks, 1)
	require.NotEmpty(t, g.Ticks[0].Snakes[0].ID)
}
