package cli

import (
	"flag"
	"fmt"
)

const (
	FlagProfile = "profile"
	FlagConfig  = "config"
)

type Flag struct {
	Name  string
	Doc   string
	Value interface{}
}

type FlagMap map[string]Flag

func (c *CLI) ParseFlags(name string, args []string) {
	// Define all flags here, then add them to the map
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	profileFlag := fs.String(FlagProfile, "", "profile to use in command")
	configFlag := fs.String(FlagConfig, "", "path to config file")
	fs.Parse(args)

	c.flags = FlagMap{
		FlagProfile: {
			Name:  FlagProfile,
			Value: *profileFlag,
		},
		FlagConfig: {
			Name:  FlagConfig,
			Value: *configFlag,
		},
	}
}

func findFlag(name string, flags FlagMap) (Flag, error) {
	if f, ok := flags[name]; ok {
		return f, nil
	}
	return Flag{}, fmt.Errorf("flag %w: %s", ErrNotFound, name)
}
