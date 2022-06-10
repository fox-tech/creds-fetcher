// Package okta implements logic for Okta's OIE flow.
package okta

import "errors"

// Client represents an Okta OIE Client used to communicate with Okta for
// authorization.
type Client struct {
	provider Provider

	id    string
	appID string
	uri   string
}

// New returns an initialized and validated Client. It returns a nil error if
// all the elements needed for a valid Client are present.
func New(id, uri string, p Provider, opts ...Option) (Client, error) {
	c := Client{
		id:       id,
		appID:    id,
		uri:      uri,
		provider: p,
	}

	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return c, err
		}
	}

	if err := c.validate(); err != nil {
		return Client{}, err
	}

	return c, nil
}

// Errors returned by New in case the parameters passed are incomplete.
var (
	ErrMissingClientConfig = errors.New("missing client ID or organization URL for Client")
	ErrNoProvider          = errors.New("no provider set")
)

func (c Client) validate() error {
	if c.id == "" || c.uri == "" {
		return ErrMissingClientConfig
	}

	if c.provider == nil {
		return ErrNoProvider
	}

	return nil
}
