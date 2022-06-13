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

func (c defaultClient) Get(getUrl string, params map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, getUrl, body)
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
