package rules

import (
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestStartSnakes(t *testing.T) {
	url := "http://not.a.snake.com/start"
	json := "{\"color\":\"#ff0000\"}"
	createClient = singleEndpointMockClient(t, url, json)

	snake := &pb.Snake{
		URL: "http://not.a.snake.com",
	}

	StartSnakes(&pb.Game{}, &pb.GameTick{
		Snakes: []*pb.Snake{snake},
	})

	require.Equal(t, "#ff0000", snake.Color)
}
