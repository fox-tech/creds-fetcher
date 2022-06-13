package cli

import (
	"errors"
	"log"
	"os"
)

const (
	defaultKey = "default"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrNoConfig  = errors.New("failed to obtain configuration")
	ErrNoCommand = errors.New("missing command name")
)

type CLI struct {
	commands       map[string]Command
	currentCommand string
	flags          FlagMap
}

func New() CLI {
	c := CLI{
		commands: map[string]Command{},
		flags:    FlagMap{},
	}

	c.AddCommand(loginCmd)
	return c
}

func (c CLI) Execute(args ...string) error {
	if err := c.ParseArguments(); err != nil {
		log.Fatalf("parsing arguments: %v", err)
		return err
	}

	command, ok := c.commands[c.currentCommand]
	if !ok {
		log.Fatal("command not found")
		return ErrNotFound
	}

	if err := command.f(c.flags); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (c *CLI) AddCommand(cmd Command) {
	c.commands[cmd.name] = cmd
}

// Parse Arguments obtains command line arguments, currently it only returns
// the command name, since it is the only one being used
func (c *CLI) ParseArguments() error {
	args := os.Args
	if len(args) == 1 {
		return ErrNoCommand
	}

	c.currentCommand = args[1]
	flags := []string{}
	if len(args) > 2 {
		flags = args[2:]
	}
	c.ParseFlags(c.currentCommand, flags)

	return nil
}
