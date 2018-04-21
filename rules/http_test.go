package rules

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	err  error
	resp func(string) *http.Response
}

func (mockHTTPClient) SetTimeout(time.Duration) {}
func (c mockHTTPClient) Get(url string) (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.resp(url), nil
}

func (c mockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.resp(url), nil
}

type readCloser struct {
	*bytes.Buffer
}

func (readCloser) Close() error {
	return nil
}

func singleEndpointMockClient(t *testing.T, url, bodyJSON string) func(time.Duration) httpClient {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString(bodyJSON)

	return func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(reqURL string) *http.Response {
				if reqURL != url {
					require.Fail(t, "invalid url")
				}
				return &http.Response{
					Body: body,
				}
			},
		}
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
