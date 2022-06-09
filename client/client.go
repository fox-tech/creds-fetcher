// Package client provides elements to interact with http
package client

import "net/http"

func NewDefault() *http.Client {
	return &http.Client{}
}
