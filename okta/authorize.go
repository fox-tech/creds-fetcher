package okta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Device represents the Device Authorization from Okta with the URL the user
// must use to sign-in the device making the credentials request. The ExpiresIn
// and Interval values are in seconds. ExpiresIn is the number in seconds that
// the values are valid and Interval is the number of seconds the device (this
// client) should wait between polling to see if the user has finished the sign
// in.
//
// More at
// https://developer.okta.com/docs/guides/device-authorization-grant/main/#request-the-device-verification-code.
type Device struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int64  `json:"expires_in"`
	Interval                int64  `json:"interval"`
}

// errorResponse represents an error returned by the Okta API.
type errorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

// PreAuthorize returns a Device Authorization with a URL to be shown to the
// user. This Device Authorization then is passed to the Authorize method to
// exchange for SSO and web SSO tokens. Returns non-nil error if there's an
// issue with the request to Okta or the parsing of Okta's response.
func (c Client) PreAuthorize() (Device, error) {
	uri := fmt.Sprintf("%s/oauth2/v1/device/authorize", c.uri)

	resp, err := http.PostForm(uri,
		url.Values{
			"client_id": []string{c.id},
			"scope":     []string{"openid okta.apps.sso"},
		},
	)
	if err != nil {
		return Device{}, fmt.Errorf("%w: %v", ErrPreAuthorizeRequest, err)
	}
	defer resp.Body.Close()

	var device Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return Device{}, fmt.Errorf("%w: %v", ErrPreAuthorizeJSONDecode, err)
	}

	return device, nil
}

// Authorize takes a Device setup and runs all the final token exchanges,
// sending the SAML assertion to the provider interface to generate credentials
// on the provider side.
func (c Client) Authorize(device Device) error {
	token, err := c.accessTokenPoll(device)
	if err != nil {
		return err
	}

	ssoToken, err := c.ssoAccessToken(token)
	if err != nil {
		return err
	}

	saml, err := c.getSAML(ssoToken)
	if err != nil {
		return err
	}

	return c.provider.GenerateCredentials(saml)
}
