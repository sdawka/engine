package rules

import (
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestCreateInitialGame(t *testing.T) {
	g, _, err := CreateInitialGame(&pb.CreateRequest{})
	require.NoError(t, err)
	require.NotNil(t, g.ID)
}

func TestCreateInitialGame_DuplicateSnakeIDs(t *testing.T) {
	_, _, err := CreateInitialGame(&pb.CreateRequest{
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
	_, frames, err := CreateInitialGame(&pb.CreateRequest{
		Width:  20,
		Height: 20,
		Snakes: []*pb.SnakeOptions{
			{ID: ""},
		},
	})
	require.NoError(t, err)
	require.Len(t, frames, 1)
	require.NotEmpty(t, frames[0].Snakes[0].ID)
}

func TestCreateInitialGame_MoreSnakesThanSpace(t *testing.T) {
	_, _, err := CreateInitialGame(&pb.CreateRequest{
		Width:  2,
		Height: 2,
		Snakes: []*pb.SnakeOptions{
			{ID: "snake_123"},
			{ID: "snake_124"},
			{ID: "snake_125"},
			{ID: "snake_126"},
			{ID: "snake_127"},
		},
	})

	require.Error(t, err)
}

func TestCreateInitialGameWithColour(t *testing.T) {
	url := setupSnakeServer(t, MoveResponse{}, StartResponse{
		Color: "#CDCDCD",
	})
	_, frame, err := CreateInitialGame(&pb.CreateRequest{
		Width:  10,
		Height: 10,
		Food:   10,
		Snakes: []*pb.SnakeOptions{
			{URL: url},
		},
	})
	require.NoError(t, err)
	require.Len(t, frame, 1)
	require.Len(t, frame[0].Snakes, 1)
	require.Equal(t, "#CDCDCD", frame[0].Snakes[0].Color)
}
