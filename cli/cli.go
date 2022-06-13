package cli

import (
	"errors"
	"flag"
	"fmt"
	"log"

	cfg "github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/configuration"
)

const (
	defaultKey = "default"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoConfig = errors.New("failed to obtain configuration")
)

type CLI struct {
	commands map[string]Command
}

func New() CLI {
	c := CLI{
		commands: map[string]Command{},
	}

	c.AddCommand(loginCmd)
	return c
}

func (c CLI) Execute() error {

	// Improvement: add handler for command arguments
	flags := c.ParseFlags()
	cmd := flag.Arg(0)

	command, ok := c.commands[cmd]
	if !ok {
		log.Fatal("command not found")
		return ErrNotFound
	}

	if err := command.f(flags); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (c CLI) AddCommand(cmd Command) {
	c.commands[cmd.name] = cmd
}

func getConfig() (cfg.Configuration, error) {
	// load configuration
	config, err := cfg.New("")
	if err != nil {
		return cfg.Configuration{}, fmt.Errorf("%w:  %v", ErrNoConfig, err)
	}
	return *config, nil
}

func getValueOrDefault(key string, values []cfg.Configuration) (cfg.Configuration, error) {
	d := cfg.Configuration{}
	dFound := false
	for i := range values {
		if values[i].OktaAppID == key {
			return values[i], nil
		}

		if values[i].OktaAppID == defaultKey {
			d = values[i]
			dFound = true
		}
	}

	if dFound {
		return d, nil
	}

	return d, fmt.Errorf("%w: value not found and default not set", ErrNotFound)
}
