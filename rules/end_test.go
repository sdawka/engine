package rules

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestNotifyGameEnd(t *testing.T) {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString("null")
	urlReceived := ""
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(url string) *http.Response {
				urlReceived = url
				return &http.Response{
					Body: body,
				}
			},
		}
	}

	snake := &pb.Snake{
		URL: "http://not.a.snake.com",
	}

	NotifyGameEnd(&pb.Game{}, &pb.GameTick{
		Snakes: []*pb.Snake{snake},
	})

	require.Equal(t, urlReceived, "http://not.a.snake.com/end")
}
