package cli

import (
	"fmt"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/aws"
)

type CommandBody func(FlagMap) error

type Command struct {
	name string
	doc  string
	f    CommandBody
}

var login = Command{
	name: "login",
	doc:  "get credentials for AWS profile",
	f:    login,
}

func login(flags FlagMap) error {
	pf, err := findFlag(FlagProfile, flags)
	if err != nil {
		return fmt.Errorf("get-credentials: %w", err)
	}

	cfg := getConfig()

	pCfg, err := getValueOrDefault(pf.Value.(string), cfg)
	if err != nil {
		return err
	}

	provider, err := aws.New(aws.Profile{
		Name:         pCfg.Name,
		RoleARN:      pCfg.AWSRoleARN,
		PrincipalARN: pCfg.AWSProviderARN,
	})

	okta, _ := okta.New(pCfg.OktaClientID, pCfg.OktaURL, aws, okta.SetAppID(appID))
	dev, err := okta.PreAuthorize()
	if err != nil {
		panic(err)
	}
	// Add message to user to follow validation on browser
	fmt.Println(dev.VerificationURIComplete)

	err = okta.Authorize(dev)
	if err != nil {
		panic(err)
	}

	return err
}
