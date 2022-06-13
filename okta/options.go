package okta

// Option represents a function that can set a configuration value to the Okta
// initialization function New. It can return a non-nil error.
type Option func(*Client) error

// SetAppID adds a different value to the app ID. This value is used for the
// scope of the web SSO token request. By default this is the same as the
// Client ID.
func SetAppID(id string) Option {
	return func(c *Client) error {
		c.appID = id
		return nil
	}
}
