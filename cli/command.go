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

	config, err := cfg.New("")
	if err != nil {
		return fmt.Errorf("%w:  %v", ErrNoConfig, err)
	}

	pCfg, err := getValueOrDefault(pf.Value.(string), []cfg.Configuration{*config})
	if err != nil {
		return err
	}

	provider, err := aws.New(aws.Profile{
		// Name:         pCfg.Name,
		Name:         "default",
		RoleARN:      pCfg.AWSRoleARN,
		PrincipalARN: pCfg.AWSProviderARN,
	})

	oktaClient, _ := okta.New(pCfg.OktaClientID, pCfg.OktaURL, provider, okta.SetAppID(pCfg.OktaAppID))
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
