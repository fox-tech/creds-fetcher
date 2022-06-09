package aws

type Option func(*Provider)

func SetProfile(p Profile) Option {
	return func(aws *Provider) {
		aws.Profile = p
	}
}

func setFileManager(fm fileSystemManager) Option {
	return func(p *Provider) {
		p.fs = fm
	}
}

func setHTTPClient(c httpClient) Option {
	return func(p *Provider) {
		p.httpClient = c
	}
}
