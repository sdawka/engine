package rules

import (
	"fmt"
	"io"
	"net/http"
	nu "net/url"
	"strings"
	"time"
)

var (
	createClient = getNetClient
)

type httpClient interface {
	SetTimeout(time.Duration)
	Get(string) (*http.Response, error)
	Post(string, string, io.Reader) (*http.Response, error)
}

type wrappedHTTPClient struct {
	*http.Client
}

func (c *wrappedHTTPClient) SetTimeout(dur time.Duration) {
	c.Client.Timeout = dur
}

func (c *wrappedHTTPClient) Get(url string) (*http.Response, error) {
	return c.Client.Get(url)
}

func (c *wrappedHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.Client.Post(url, contentType, body)
}

func getNetClient(duration time.Duration) httpClient {
	return &wrappedHTTPClient{
		Client: &http.Client{
			Timeout: duration,
		},
	}
}

func isValidURL(url string) bool {
	if len(url) == 0 {
		return false
	}

	parsed, err := nu.Parse(url)
	if err != nil {
		return false
	}

	if len(parsed.Scheme) == 0 {
		return false
	}

	return true
}

func cleanURL(url string) string {
	if !strings.HasSuffix(url, "/") {
		return fmt.Sprintf("%s/", url)
	}
	return url
}

func getURL(url, path string) string {
	u := cleanURL(url)
	return fmt.Sprintf("%s%s", u, path)
}
