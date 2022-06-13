package cli

import (
	"fmt"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/aws"
	cfg "github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/configuration"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/okta"
)

type CommandBody func(FlagMap) error

type Command struct {
	name string
	doc  string
	f    CommandBody
}

var loginCmd = Command{
	name: "login",
	doc:  "get credentials for AWS profile",
	f:    login,
}

func login(flags FlagMap) error {
	pf, err := findFlag(FlagProfile, flags)
	if err != nil {
		return fmt.Errorf("login : %w", err)
	}

	cf, err := findFlag(FlagConfig, flags)
	if err != nil {
		return fmt.Errorf("login : %w", err)
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
		panic(err)
	}
	// Add message to user to follow validation on browser
	fmt.Println(dev.VerificationURIComplete)

	err = oktaClient.Authorize(dev)
	if err != nil {
		panic(err)
	}

	return err
}
