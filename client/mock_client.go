package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type MockHttpClient struct {
	GetStatusCode  int
	GetStatus      string
	GetBodyData    []byte
	GetErr         error
	PostStatusCode int
	PostStatus     string
	PostBodyData   []byte
	PostErr        error
}

func (m MockHttpClient) Get(r_url string, params map[string]string, body io.Reader) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.GetStatusCode,
		Status:     fmt.Sprintf("%d : %s", m.GetStatusCode, m.GetStatus),
		Body:       io.NopCloser(bytes.NewBuffer(m.GetBodyData)),
		Request: &http.Request{
			Method: http.MethodGet,
		},
	}, m.GetErr
}

func (m MockHttpClient) Post(postUrl string, params map[string]string, headers map[string]string, body interface{}) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.PostStatusCode,
		Status:     fmt.Sprintf("%d : %s", m.PostStatusCode, m.PostStatus),
		Body:       io.NopCloser(bytes.NewBuffer(m.PostBodyData)),
		Request: &http.Request{
			Method: http.MethodPost,
		},
	}, m.PostErr
}
