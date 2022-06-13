// Package okta implements logic for Okta's OIE flow.
package okta

import "errors"

var (
	// Errors returned by New in case the parameters passed are incomplete.
	ErrMissingClientConfig = errors.New("missing client ID or organization URL for Client")
	ErrNoProvider          = errors.New("no provider set")

	// ErrPreAuthorizeJSONDecode is returned when the JSON response from
	// PreAuthorize failed to decode.
	ErrPreAuthorizeJSONDecode = errors.New("json decode")

	// ErrPreAuthorizeRequest is returned when the PreAuthorize request failed.
	ErrPreAuthorizeRequest = errors.New("preauthorize request")

	// ErrSAMLRequest is returned when there's a failure related to the request.
	// The response is not the expected response or it failed to complete the
	// request.
	ErrSAMLRequest = errors.New("saml request")

	// ErrSAMLResponse is returned when there's a failure reading the response from
	// the request.
	ErrSAMLResponse = errors.New("saml response")

	// ErrNoSAMLResponseFound is returned when the HTML from the SSO login has no
	// SAMLResponse tag in its body.
	ErrNoSAMLResponseFound = errors.New("no SAMLResponse found")

	// ErrParsingHTML is returned when the HTML from the SSO login can't be parsed.
	// This probably will never happen because the HTML parser returns early on any
	// error, without error. And the only error that handles is io.EOF, as an exit
	// as well.
	ErrParsingHTML = errors.New("parsing html")

	// ErrDeviceAuthorizationExpired is returned when the Device Authorization
	// credentials expired. In this case, the PreAuthorize has to be run again and
	// the user will be provided with a new verification URL.
	ErrDeviceAuthorizationExpired = errors.New("device authorization expired")

	// ErrAccessTokenRequest is returned when the server fails to fulfill the
	// device code authorization exchange request or the response is different to
	// http.StatusOK and http.StatusBadRequest.
	ErrAccessTokenRequest = errors.New("accessToken request")

	// ErrAccessTokenJSONDecode is returned when the JSON response from the device
	// code authorization exchange can't be properly decoded.
	ErrAccessTokenJSONDecode = errors.New("json decode")

	// ErrSSOJSONDecode is returned when the response for the web SSO token can't
	// be decoded into JSON.
	ErrSSOJSONDecode = errors.New("sso json decode")

	// ErrSSORequest is returned when the server fails to fulfill the SSO token
	// request or the response is different to http.StatusOK.
	ErrSSORequest = errors.New("sso request")
)

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

func (c Client) validate() error {
	if c.id == "" || c.uri == "" {
		return ErrMissingClientConfig
	}

	if c.provider == nil {
		return ErrNoProvider
	}

	return nil
}
