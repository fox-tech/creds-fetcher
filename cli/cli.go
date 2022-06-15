package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	defaultKey = "default"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrNoCommand = errors.New("missing command name")
)

// CLI represents an interpreter that will execute a given command
type CLI struct {
	// args represents the command line arguments sent to the CLI. They are set
	// to os.Args when New() is called
	args []string
	// commands represents the available commands to the CLI
	commands CommandMap
	// flags represents the flags sent to the CLI. They are parsed based on args
	flags FlagMap
}

// New creates a CLI instance with default values and adds the
// supported commands: login
func New() CLI {
	c := CLI{
		commands: CommandMap{},
		flags:    FlagMap{},
	}
	c.args = os.Args
	c.AddCommand(loginCmd)
	return c
}

// Execute executes a given command by parsing the command line arguments
func (c CLI) Execute(args ...string) error {
	cmdName, err := c.ParseArguments()
	if err != nil {
		err = fmt.Errorf("parsing arguments: %w", err)
		log.Print(err)
		return err
	}

	command, ok := c.commands[cmdName]
	if !ok {
		err := fmt.Errorf("command %s %w", cmdName, ErrNotFound)
		log.Print(err)
		return err
	}

	if err := command.f(c.flags); err != nil {
		err = fmt.Errorf("%s: %w", cmdName, err)
		log.Print(err)
		return err
	}

	return nil
}

// AddCommand adds a new supported command to the CLI
func (c *CLI) AddCommand(cmd Command) {
	c.commands[cmd.name] = cmd
}

// Parse Arguments obtains command line arguments and parses flags,
// currently it only returns the command name
func (c *CLI) ParseArguments() (string, error) {
	if len(c.args) == 1 {
		return "", ErrNoCommand
	}

	// verifying the command name is not a flag
	cmdName := c.args[1]
	if strings.HasPrefix(cmdName, "-") {
		return "", ErrNoCommand
	}

	flags := []string{}
	if len(c.args) > 2 {
		flags = c.args[2:]
	}
	c.ParseFlags(cmdName, flags)

	return cmdName, nil
}
