package rules

import (
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestBuildSnakeRequest(t *testing.T) {
	req := buildSnakeRequest(&pb.Game{
		ID: "game_123",
	}, &pb.GameFrame{
		Snakes: []*pb.Snake{
			{
				ID: "snake_123",
				Body: []*pb.Point{
					{X: 1, Y: 1},
				},
			},
		},
	}, "snake_123")
	require.Equal(t, "game_123", req.Game.ID)
	require.Equal(t, []Coords{{X: 1, Y: 1}}, req.Board.Snakes[0].Body)
	require.Equal(t, []Coords{{X: 1, Y: 1}}, req.You.Body)
}
