package configuration

import "errors"

var (
	ErrNoConfiguration              = errors.New("no configuration file found")
	ErrEmptyConfigurationFile       = errors.New("invalid configuration file, cannot be empty")
	ErrCannotParseConfigurationFile = errors.New("unable to parse configuration file")
	ErrInvalidAWSProviderARN        = errors.New("invalid aws_provider_arn, cannot be empty")
	ErrInvalidAWSRoleARN            = errors.New("invalid aws_role_arn, cannot be empty")
	ErrInvalidOktaClientID          = errors.New("invalid okta_client_id, cannot be empty")
	ErrInvalidOktaAppID             = errors.New("invalid okta_app_id, cannot be empty")
	ErrInvalidOktaURL               = errors.New("invalid okta_url, cannot be empty")
)

var (
	sources = []string{
		"./config.json",
		"./config.toml",
		"stdin",
	}

	decoders = []decoder{
		decodeAsJSON,
		decodeAsTOML,
	}
)

// New will create a new Configuration from the configuration local configuration file.
// An optional override location argument allows for configurations located outside of the
// default scope. The default scope is as follows:
//	- ./configuration.json
//	- ./configuration.toml
//	- stdin
func New(overrideLocation string) (cfg *Configuration, err error) {
	if cfg, err = getConfiguration(overrideLocation); err != nil {
		return
	}

	if err = cfg.Validate(); err != nil {
		cfg = nil
		return
	}

	return
}

type Configuration struct {
	AWSProviderARN string `toml:"aws_provider_arn" json:"aws_provider_arn" env:"aws_provider_arn"`
	AWSRoleARN     string `toml:"aws_role_arn" json:"aws_role_arn" env:"aws_role_arn"`
	OktaClientID   string `toml:"okta_client_id" json:"okta_client_id" env:"okta_client_id"`
	OktaAppID      string `toml:"okta_app_id" json:"okta_app_id" env:"okta_app_id"`
	OktaURL        string `toml:"okta_url" json:"okta_url" env:"okta_url"`
}

func (c *Configuration) OverrideWith(in *Configuration) {
	if len(in.AWSProviderARN) > 0 {
		c.AWSProviderARN = in.AWSProviderARN
	}

	if len(in.AWSRoleARN) > 0 {
		c.AWSRoleARN = in.AWSRoleARN
	}

	if len(in.OktaClientID) > 0 {
		c.OktaClientID = in.OktaClientID
	}

	if len(in.OktaAppID) > 0 {
		c.OktaAppID = in.OktaAppID
	}

	if len(in.OktaURL) > 0 {
		c.OktaURL = in.OktaURL
	}
}

func (c *Configuration) Validate() (err error) {
	if len(c.AWSProviderARN) == 0 {
		return ErrInvalidAWSProviderARN
	}

	if len(c.AWSRoleARN) == 0 {
		return ErrInvalidAWSRoleARN
	}

	if len(c.OktaClientID) == 0 {
		return ErrInvalidOktaClientID
	}

	if len(c.OktaAppID) == 0 {
		return ErrInvalidOktaAppID
	}

	if len(c.OktaURL) == 0 {
		return ErrInvalidOktaURL
	}

	return
}
