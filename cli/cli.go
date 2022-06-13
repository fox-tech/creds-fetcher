package cli

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

const (
	defaultKey = "default"
)

var (
	ErrNotFound = errors.New("not found")
)

type CLI struct {
	commands map[string]Command
}

func New() CLI {
	c := CLI{
		commands: map[string]Command{},
	}

	c.AddCommand(login)
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

func getConfig() []Configuration {
	// load configuration
	// cfg, err := New("./path/to/config/config.json")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return []Configuration{
		{
			AWSProviderARN: "provider-arn",
			AWSRoleARN:     "role-are",
			Name:           "test",
		},
	}
}

func getValueOrDefault(key string, values []Configuration) (Configuration, error) {
	d := Configuration{}
	dFound := false
	for i := range values {
		if values[i].Name == key {
			return values[i], nil
		}

		if values[i].Name == defaultKey {
			d = values[i]
			dFound = true
		}
	}

	if dFound {
		return d, nil
	}

	return d, fmt.Errorf("%w: value not found and default not set", ErrNotFound)
}
