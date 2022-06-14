package cli

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func Test_New(t *testing.T) {
	expect := CLI{
		commands: CommandMap{loginCmd.name: loginCmd},
		flags:    FlagMap{},
	}

	got := New()
	if _, ok := got.commands[loginCmd.name]; !ok {
		t.Errorf("New() expected commands: %v, got: %v", expect.commands, got.commands)
	}
	if !reflect.DeepEqual(expect.flags, got.flags) {
		t.Errorf("New() expected flags: %v, got: %v", expect.flags, got.flags)
	}
}

func Test_ParseArguments(t *testing.T) {
	type expect struct {
		cmdName string
		flags   FlagMap
		err     error
	}

	tests := []struct {
		name string
		args string
		expect
	}{
		{
			name: "parse command name and flags",
			args: "cli login -profile dev1 -config=.aws/config",
			expect: expect{
				cmdName: "login",
				flags: FlagMap{
					FlagProfile: {Name: FlagProfile, Value: "dev1"},
					FlagConfig:  {Name: FlagConfig, Value: ".aws/config"},
				},
				err: nil,
			},
		},
		{
			name: "error: no given command",
			args: "",
			expect: expect{
				err: ErrNoCommand,
			},
		},
		{
			name: "error: no given command with flags",
			args: "cli -profile dev1",
			expect: expect{
				err: ErrNoCommand,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CLI{
				args:     strings.Split(tt.args, " "),
				commands: CommandMap{},
			}

			cmdName, err := c.ParseArguments()

			if tt.expect.cmdName != cmdName {
				t.Errorf("ParseArguments() expected command name: %v, got %v", tt.expect.cmdName, cmdName)
			}

			if !reflect.DeepEqual(tt.expect.flags, c.flags) {
				t.Errorf("ParseArguments() expected flags: %v, got %v", tt.expect.flags, c.flags)
			}

			if tt.expect.err != err {
				t.Errorf("ParseArguments() expected error: %v, got %v", tt.expect.err, err)
			}
		})
	}
}

func Test_Execute(t *testing.T) {
	c := CLI{
		commands: CommandMap{
			"test": Command{
				name: "test",
				doc:  "testing cli",
				f: func(flags FlagMap) error {
					return nil
				},
			},
			"testError": Command{
				name: "test",
				doc:  "testing cli",
				f: func(flags FlagMap) error {
					return ErrNoConfig
				},
			},
		},
	}

	tests := []struct {
		name   string
		args   string
		expect error
	}{
		{
			name:   "test success command",
			args:   "cli test",
			expect: nil,
		},
		{
			name:   "error: parsing arguments",
			args:   "cli",
			expect: ErrNoCommand,
		},
		{
			name:   "error: unknown command",
			args:   "cli testing",
			expect: ErrNotFound,
		},
		{
			name:   "error: error in command execution",
			args:   "cli testError",
			expect: ErrNoConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.args = strings.Split(tt.args, " ")
			err := c.Execute()

			if !errors.Is(err, tt.expect) {
				t.Errorf("Execute() expected error: %v, got %v", tt.expect, err)
			}
		})
	}
}
