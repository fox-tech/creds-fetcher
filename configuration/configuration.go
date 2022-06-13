package configuration

import (
	"errors"
	"fmt"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/env"
)

var (
	ErrNoConfiguration              = errors.New("no configuration file found")
	ErrEmptyConfigurationFile       = errors.New("invalid configuration file, cannot be empty")
	ErrCannotParseConfigurationFile = errors.New("unable to parse configuration file")
	ErrInvalidAWSProviderARN        = errors.New("invalid aws_provider_arn, cannot be empty")
	ErrInvalidAWSRoleARN            = errors.New("invalid aws_role_arn, cannot be empty")
	ErrInvalidOktaClientID          = errors.New("invalid okta_client_id, cannot be empty")
	ErrInvalidOktaAppID             = errors.New("invalid okta_app_id, cannot be empty")
	ErrInvalidOktaURL               = errors.New("invalid okta_url, cannot be empty")
	ErrNilReader                    = errors.New("invalid reader, cannot be nil")
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
//
// Once the initial configuration is loaded, the library will then check the os environment
// variables. In order to determine if any variable overrides are required. If so, the provided
// environment values which are set will be applied to the loaded configuration.
func New(profile, overrideLocation string) (cfg *Configuration, err error) {
	var cfgs map[string]*Configuration
	if cfgs, err = getConfigurations(overrideLocation); err != nil {
		return
	}

	var ok bool
	if cfg, ok = cfgs[profile]; !ok {
		err = fmt.Errorf("profile of <%s> not found", profile)
		return
	}

	if err = cfg.Validate(); err != nil {
		cfg = nil
		return
	}

	var overrideValues Configuration
	if err = env.Unmarshal(&overrideValues); err != nil {
		err = fmt.Errorf("error Unmarshaling environment variables: %v", err)
		return
	}

	cfg.OverrideWith(&overrideValues)
	return
}

type Configuration struct {
	AWSProviderARN string `toml:"aws_provider_arn" json:"aws_provider_arn" env:"AWS_PROVIDER_ARN"`
	AWSRoleARN     string `toml:"aws_role_arn" json:"aws_role_arn" env:"AWS_ROLE_ARN"`
	OktaClientID   string `toml:"okta_client_id" json:"okta_client_id" env:"OKTA_CLIENT_ID"`
	OktaAppID      string `toml:"okta_app_id" json:"okta_app_id" env:"OKTA_APP_ID"`
	OktaURL        string `toml:"okta_url" json:"okta_url" env:"OKTA_URL"`
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
