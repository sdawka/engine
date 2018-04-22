package rules

import (
	"errors"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestStartSnakes(t *testing.T) {
	resetPalette(defaultColors)
	snake := getSnakeAfterStart(t, "{\"color\":\"#ff0000\"}", 200)
	require.Equal(t, "#ff0000", snake.Color)
}

func TestStartSnakesMissingColor(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})
	snake := getSnakeAfterStart(t, "{}", 200)
	require.Equal(t, "red", snake.Color)
}

func TestStartSnakesMissingEndpoint(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})
	snake := getSnakeAfterStart(t, "{}", 404)
	require.Equal(t, "red", snake.Color)
}

func TestStartSnakesMissingServer(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})
	snake := getSnakeAfterMissingServer(t)
	require.Equal(t, "red", snake.Color)
}

func getSnakeAfterStart(t *testing.T, json string, statusCode int) *pb.Snake {
	url := "http://good-server/start"
	createClient = singleEndpointMockClient(t, url, json, statusCode)

	snake := &pb.Snake{
		URL: "http://good-server",
	}

	NotifyGameStart(&pb.Game{}, &pb.GameTick{
		Snakes: []*pb.Snake{snake},
	})

	return snake
}

func getSnakeAfterMissingServer(t *testing.T) *pb.Snake {
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			err: errors.New("fail"),
		}
	}

	snake := &pb.Snake{
		URL: "http://dead-server",
	}

	NotifyGameStart(&pb.Game{}, &pb.GameTick{
		Snakes: []*pb.Snake{snake},
	})

	return snake
}
