package cli

import (
	"flag"
	"fmt"
)

const (
	FlagProfile = "profile"
)

type Flag struct {
	Name  string
	Doc   string
	Value interface{}
}

type FlagMap map[string]Flag

func (c CLI) ParseFlags() FlagMap {
	// Define all flags here, then add them to the map
	pFlag := flag.String(FlagProfile, "", "profile to use in command")
	flag.Parse()

	return map[string]Flag{
		FlagProfile: {
			Name:  FlagProfile,
			Value: *pFlag,
		},
	}
}

func findFlag(name string, flags FlagMap) (Flag, error) {
	if f, ok := flags[name]; ok {
		return f, nil
	}
	return Flag{}, fmt.Errorf("flag %w: %s", ErrNotFound, name)
}
