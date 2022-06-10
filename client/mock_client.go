package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type MockHttpClient struct {
	GetStatusCode int
	GetStatus     string
	GetBodyData   []byte
	GetErr        error
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
