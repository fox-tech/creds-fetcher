package cli

import (
	"errors"
	"fmt"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/aws"
	cfg "github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/configuration"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/okta"
)

var (
	ErrNoConfig             = errors.New("failed to obtain configuration")
	ErrAuthenticationFailed = errors.New("failed to authenticate")
)

var loginCmd = Command{
	name: "login",
	doc:  " get credentials for AWS profile",
	f:    login,
}

// login uses OKTA to authorize and get credentials from AWS STS
func login(flags FlagMap) error {
	pf, err := findFlag(FlagProfile, flags)
	if err != nil {
		return err
	}

	cf, err := findFlag(FlagConfig, flags)
	if err != nil {
		return err
	}

	profName := pf.Value.(string)
	if profName == "" {
		profName = defaultKey
	}

	configFile := cf.Value.(string)

	config, err := cfg.New(profName, configFile)
	if err != nil {
		return fmt.Errorf("%w:  %v", ErrNoConfig, err)
	}

	provider, err := aws.New(aws.Profile{
		Name:         profName,
		RoleARN:      config.AWSRoleARN,
		PrincipalARN: config.AWSProviderARN,
	})

	oktaClient, _ := okta.New(config.OktaClientID, config.OktaURL, provider, okta.SetAppID(config.OktaAppID))

	dev, err := oktaClient.PreAuthorize()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}
	fmt.Println("Open URL and follow authentication in browser")
	fmt.Println(dev.VerificationURIComplete)

	err = oktaClient.Authorize(dev)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	return nil
}
