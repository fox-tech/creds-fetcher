package aws

// Option represents an optional configuration value passed to the
// provider object to change the default value set during initialization.
type Option func(*Provider)

// setFileManager returns a function to assign the passed fileSystemManager
// to the provider.
func setFileManager(fm fileSystemManager) Option {
	return func(p *Provider) {
		p.fs = fm
	}
}

// setHTTPClient returns a function to assign the passed httpClient
// to the provider.
func setHTTPClient(c httpClient) Option {
	return func(p *Provider) {
		p.httpClient = c
	}
}
