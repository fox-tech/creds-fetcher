package cli

import (
	"errors"
	"reflect"
	"testing"
)

func Test_ParseFlags(t *testing.T) {
	type args struct {
		name string
		args []string
	}

	tests := []struct {
		name   string
		expect FlagMap
		args
	}{
		{
			name: "parse profile flag with '-flag value' format",
			args: args{
				name: "test",
				args: []string{"-profile", "dev1"},
			},
			expect: FlagMap{
				FlagProfile: {Name: FlagProfile, Value: "dev1"},
				FlagConfig:  {Name: FlagConfig, Value: ""},
			},
		},
		{
			name: "parse profile and config flags with '-flag value' format",
			args: args{
				name: "test",
				args: []string{"-profile", "dev1", "-config", ".aws/config"},
			},
			expect: FlagMap{
				FlagProfile: {Name: FlagProfile, Value: "dev1"},
				FlagConfig:  {Name: FlagConfig, Value: ".aws/config"},
			},
		},
		{
			name: "parse profile and config flags with '-flag=value' format",
			args: args{
				name: "test",
				args: []string{"-profile=dev1", "-config=.aws/config"},
			},
			expect: FlagMap{
				FlagProfile: {Name: FlagProfile, Value: "dev1"},
				FlagConfig:  {Name: FlagConfig, Value: ".aws/config"},
			},
		},
		{
			name: "parse profile and config flags with mixed format",
			args: args{
				name: "test",
				args: []string{"-profile", "dev1", "-config=.aws/config"},
			},
			expect: FlagMap{
				FlagProfile: {Name: FlagProfile, Value: "dev1"},
				FlagConfig:  {Name: FlagConfig, Value: ".aws/config"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.ParseFlags(tt.args.name, tt.args.args)

			if !reflect.DeepEqual(tt.expect, c.flags) {
				t.Errorf("ParseFlags() expected: %v, got %v", tt.expect, c.flags)
			}
		})
	}
}

func Test_findFlag(t *testing.T) {
	type expect struct {
		flag Flag
		err  error
	}

	flags := FlagMap{
		"flag1": {
			Name:  "flag1",
			Value: "abc",
		},
		"flag2": {
			Name:  "flag2",
			Value: "def",
		},
	}

	tests := []struct {
		name  string
		key   string
		flags FlagMap
		expect
	}{
		{
			name:  "success: flag is found",
			key:   "flag1",
			flags: flags,
			expect: expect{
				err:  nil,
				flag: Flag{Name: "flag1", Value: "abc"},
			},
		},
		{
			name:  "error: flag is not found",
			key:   "flagX",
			flags: flags,
			expect: expect{
				err:  ErrNotFound,
				flag: Flag{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := findFlag(tt.key, tt.flags)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("findFlag() expected error: %s, got: %s", tt.expect.err, err)
			}

			if !reflect.DeepEqual(tt.expect.flag, f) {
				t.Errorf("findFlag() expected flag: %v, got: %v", tt.expect.flag, f)
			}
		})
	}
}

func Test_IsEqual(t *testing.T) {

}
