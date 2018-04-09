package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/battlesnakeio/engine/controller/pb"
)

func TestCheckForGameOver_SinglePlayer(t *testing.T) {

	gameTick := &pb.GameTick{
		Snakes: []*pb.Snake{
			{Death: &pb.Death{}},
		},
	}
	res := CheckForGameOver(GameModeSinglePlayer, gameTick)
	require.True(t, res)

	gameTick.Snakes[0].Death = nil
	res = CheckForGameOver(GameModeSinglePlayer, gameTick)
	require.False(t, res)
}
