package aws

type Option func(*Provider)

func SetProfile(p Profile) Option {
	return func(aws *Provider) {
		aws.Profile = p
	}
}
