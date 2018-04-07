package rules

import (
	"bytes"
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

type readCloser struct {
	*bytes.Buffer
}

func (readCloser) Close() error {
	return nil
}

func TestGatherSnakeMoves(t *testing.T) {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString("{\"move\":\"up\"}")
	mock := mockHTTPClient{
		resp: &http.Response{
			Body: body,
		},
	}
	netClient = mock
	updates := GatherSnakeMoves(1*time.Second, &pb.GameTick{
		Snakes: []*pb.Snake{
			&pb.Snake{
				URL: "http://not.a.snake.com",
			},
		},
	})
	select {
	case update := <-updates:
		require.Equal(t, "up", update.Move)
	case <-time.After(250 * time.Millisecond):
		require.Fail(t, "No update received over updates channel")
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		URL      string
		Expected bool
	}{
		{URL: "", Expected: false},
		{URL: "aksdjflaskjd", Expected: false},
		{URL: "http://127.0.0.1:8001", Expected: true},
		{URL: "https://snake.battlesnake.io/something/something", Expected: true},
	}

	for _, test := range tests {
		actual := isValidURL(test.URL)
		require.Equal(t, test.Expected, actual, "URL: %s", test.URL)
	}
}
