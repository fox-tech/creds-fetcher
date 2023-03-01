package cli

import "github.com/fox-tech/creds-fetcher/okta"

type CommandAction func(FlagMap) error

type CommandMap map[string]Command

type Command struct {
	name string
	doc  string
	f    CommandAction

	authenticator Authenticator
}

type Authenticator interface {
	PreAuthorize() (okta.Device, error)
	Authorize(device okta.Device) error
}
