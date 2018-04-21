package rules

import (
	"encoding/json"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
)

// SnakeMetadata contains a snake and the metadata sent in the start response
// from the snake server.
type SnakeMetadata struct {
	Snake *pb.Snake
	Color string
	Err   error
}

func toSnakeStartResponse(resp snakeResponse) SnakeMetadata {
	if resp.err == nil {
		startResponse := StartResponse{}
		err := json.Unmarshal(resp.data, &startResponse)
		if err != nil {
			return SnakeMetadata{
				Snake: resp.snake,
				Err:   err,
			}
		}
		return SnakeMetadata{
			Snake: resp.snake,
			Color: startResponse.Color,
		}
	}

	return SnakeMetadata{
		Snake: resp.snake,
		Err:   resp.err,
	}
}

// StartSnakes calls /start on every snake and then adds metadata from the
// response to the pb.Snake object.
func StartSnakes(game *pb.Game, startState *pb.GameTick) {
	// Be nice and give snake servers a long time to respond to /start in case
	// it's a sleeping heroku dyno or something like that.
	timeout := 15 * time.Second
	responses := gatherSnakeStartResponses(timeout, game, startState)

	for _, resp := range responses {
		if resp.Err != nil {
			resp.Snake.Death = &pb.Death{
				Cause: DeathCauseStartFail,
				Turn:  0,
			}
		} else {
			resp.Snake.Color = resp.Color
		}
	}
}

func gatherSnakeStartResponses(timeout time.Duration, game *pb.Game, startState *pb.GameTick) []SnakeMetadata {
	responses := gatherAllSnakeResponses("start", timeout, game, startState)

	ret := []SnakeMetadata{}
	for _, resp := range responses {
		ret = append(ret, toSnakeStartResponse(resp))
	}
	return ret
}
