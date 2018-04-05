package rules

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	err  error
	resp *http.Response
}

func (mockHTTPClient) SetTimeout(time.Duration) {}
func (c mockHTTPClient) Get(url string) (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.resp, nil
}

func TestGatherSnakeMoves(t *testing.T) {
	mock := mockHTTPClient{
		resp: &http.Response{},
	}
	netClient = mock
	updates := GatherSnakeMoves(&pb.Game{
		Snakes: []*pb.Snake{
			&pb.Snake{
				URL: "http://not.a.snake.com",
			},
		},
	})
	select {
	case update := <-updates:
		// do something
		fmt.Println(update)
	case <-time.After(250 * time.Millisecond):
		require.Fail(t, "No update received over updates channel")
	}
}
