package rules

import (
	"net/http"
	"time"
)

type httpClient interface {
	SetTimeout(time.Duration)
	Get(string) (*http.Response, error)
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
