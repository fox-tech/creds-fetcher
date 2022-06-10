// Package client provides elements to interact with http
package client

import (
	"io"
	"net/http"
	"net/url"
)

type defaultClient struct {
	cl *http.Client
}

func NewDefault() defaultClient {
	return defaultClient{
		cl: &http.Client{},
	}
}

func (c defaultClient) Get(r_url string, params map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, r_url, body)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return c.cl.Do(req)
}
