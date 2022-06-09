package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type mockHttpClient struct {
	response *http.Response
	err      error
}

func NewMock(sc int, st string, data []byte, err error) mockHttpClient {
	resp := http.Response{
		StatusCode: sc,
		Status:     fmt.Sprintf("%d : %s", sc, st),
		Body:       io.NopCloser(bytes.NewBuffer(data)),
	}
	return mockHttpClient{
		response: &resp,
		err:      err,
	}
}

func (m mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}
