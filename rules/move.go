package rules

import (
	"encoding/json"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
)

// SnakeUpdate bundles together a snake with a move for processing
type SnakeUpdate struct {
	Snake *pb.Snake
	Move  string
	Err   error
}

func toSnakeUpdate(resp snakeResponse) *SnakeUpdate {
	if resp.err == nil {
		moveResponse := MoveResponse{}
		err := json.Unmarshal(resp.data, &moveResponse)
		if err != nil {
			return &SnakeUpdate{
				Snake: resp.snake,
				Err:   err,
			}
		}
		return &SnakeUpdate{
			Snake: resp.snake,
			Move:  moveResponse.Move,
		}
	}

	return &SnakeUpdate{
		Snake: resp.snake,
		Err:   resp.err,
	}
}

// GatherSnakeMoves goes and queries each snake for the snake move
func GatherSnakeMoves(timeout time.Duration, game *pb.Game, gameTick *pb.GameTick) []*SnakeUpdate {
	responses := gatherAliveSnakeResponses("move", timeout, game, gameTick)

	ret := []*SnakeUpdate{}
	for _, resp := range responses {
		ret = append(ret, toSnakeUpdate(resp))
	}
	return ret
}
