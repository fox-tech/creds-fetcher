package client

import "net/http"

func NewDefault() *http.Client {
	return &http.Client{}
}
