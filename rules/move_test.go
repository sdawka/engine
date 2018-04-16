package rules

import (
	"bytes"
	"io"
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

func (c mockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
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
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: &http.Response{
				Body: body,
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

func TestGetURL(t *testing.T) {
	tests := []struct {
		URL      string
		Path     string
		Expected string
	}{
		{URL: "http://localhost", Path: "move", Expected: "http://localhost/move"},
		{URL: "http://localhost/", Path: "move", Expected: "http://localhost/move"},
	}

	for _, test := range tests {
		actual := getURL(test.URL, test.Path)
		require.Equal(t, test.Expected, actual)
	}
}
