package rules

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestGatherSnakeMoves(t *testing.T) {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString("{\"move\":\"up\"}")
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(url string) *http.Response {
				if url != "http://not.a.snake.com/move" {
					require.Fail(t, "invalid url")
				}
				return &http.Response{
					Body: body,
				}
			},
		}
	}
	updates := make(chan *SnakeUpdate)
	go func() {
		u := GatherSnakeMoves(1*time.Second, &pb.Game{}, &pb.GameTick{
			Snakes: []*pb.Snake{
				&pb.Snake{
					URL: "http://not.a.snake.com",
				},
			},
		})
		if len(u) > 0 {
			updates <- u[0]
		}
	}()
	select {
	case update := <-updates:
		require.Equal(t, "up", update.Move)
	case <-time.After(250 * time.Millisecond):
		require.Fail(t, "No update received over updates channel")
	}
}
