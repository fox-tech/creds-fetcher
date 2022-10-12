// Package client provides elements to interact with http
package client

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type defaultClient struct {
	cl *http.Client
}

var (
	ErrInvalidBody = errors.New("invalid body for request")
)

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

func (c defaultClient) Post(postUrl string, params map[string]string, headers map[string]string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	switch body.(type) {
	//x-www-form-urlencoded values
	case map[string]string:
		q := url.Values{}
		for k, v := range body.(map[string]string) {
			q.Add(k, v)
		}
		encodedValues := q.Encode()
		reqBody = strings.NewReader(encodedValues)
	case io.Reader:
		reqBody = body.(io.Reader)
	default:
		return nil, ErrInvalidBody
	}

	req, err := http.NewRequest(http.MethodPost, postUrl, reqBody)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return c.cl.Do(req)
}
