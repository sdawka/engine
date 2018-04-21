package rules

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestStartSnakes(t *testing.T) {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString("{\"color\":\"#ff0000\"}")
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(url string) *http.Response {
				if url != "http://not.a.snake.com/start" {
					require.Fail(t, "invalid url")
				}
				return &http.Response{
					Body: body,
				}
			},
		}
	}

	snake := &pb.Snake{
		URL: "http://not.a.snake.com",
	}

	StartSnakes(&pb.Game{}, &pb.GameTick{
		Snakes: []*pb.Snake{snake},
	})

	require.Equal(t, "#ff0000", snake.Color)
}
