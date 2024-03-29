package configuration

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

const (
	exampleJSON = `
{
	"my_profile" : {
		"aws_provider_arn" : "1",
		"aws_role_arn" : "2",
		"okta_client_id" : "3",
		"okta_app_id" : "4",
		"okta_url" : "5"
	}			
}
`

	exampleJSONInvalid = `
{
	"my_profile" : {
		"aws_role_arn" : "2",
		"okta_client_id" : "3",
		"okta_app_id" : "4",
		"okta_url" : "5"				
	}
}
`
	exampleJSONArray = `["hello world", "foo", "bar", "baz"]`

	exampleTOML = `
[my_profile]
aws_provider_arn = "1"
aws_role_arn = "2"
okta_client_id = "3"
okta_app_id = "4"
okta_url = "5"
`
)

var (
	exampleConfiguration = &Configuration{
		AWSProviderARN: "1",
		AWSRoleARN:     "2",
		OktaClientID:   "3",
		OktaAppID:      "4",
		OktaURL:        "5",
	}

	exampleConfigurations = map[string]*Configuration{
		"my_profile": exampleConfiguration,
	}
)

func TestNew(t *testing.T) {
	type args struct {
		profile          string
		overrideLocation string
	}

	tests := []struct {
		name string
		args args
		prep func() (toRemove *os.File, err error)

		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				profile: "my_profile",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSON)
				os.Stdin = toRemove
				return
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "success (override location)",
			args: args{
				profile:          "my_profile",
				overrideLocation: "./Test_New.override.json",
			},
			prep: func() (tmp *os.File, err error) {
				return createTestFile("./Test_New.override.json", exampleJSON)
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "success (with override values)",
			args: args{
				profile: "my_profile",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSON)
				os.Stdin = toRemove
				os.Setenv("AWS_PROVIDER_ARN", "1n")
				os.Setenv("AWS_ROLE_ARN", "2n")
				return
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1n",
				AWSRoleARN:     "2n",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name: "failure (invalid configuration)",
			args: args{
				profile: "my_profile",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSONInvalid)
				os.Stdin = toRemove
				return
			},
			wantErr: true,
		},
		{
			name: "failure (missing profile)",
			args: args{
				profile: "",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSON)
				os.Stdin = toRemove
				return
			},
			wantErr: true,
		},
		{
			name: "failure (closed file)",
			args: args{
				profile: "my_profile",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestClosedTempfile()
				os.Stdin = toRemove
				return
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				toRemove *os.File
				err      error
			)

			if toRemove, err = tt.prep(); err != nil {
				t.Errorf("New() error preparing test: %v", err)
				return
			}

			if toRemove != nil {
				defer os.Remove(toRemove.Name())
				defer toRemove.Close()
			}

			gotCfg, err := New(tt.args.profile, tt.args.overrideLocation)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("New() = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
	}
}

func TestConfiguration_OverrideWith(t *testing.T) {
	type fields struct {
		AWSProviderARN string
		AWSRoleARN     string
		OktaClientID   string
		OktaAppID      string
		OktaURL        string
	}

	type args struct {
		in *Configuration
	}

	baseFields := fields{
		AWSProviderARN: "1",
		AWSRoleARN:     "2",
		OktaClientID:   "3",
		OktaAppID:      "4",
		OktaURL:        "5",
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantCfg *Configuration
	}{
		{
			name:   "AWS Provider ARN",
			fields: baseFields,
			args: args{
				in: &Configuration{
					AWSProviderARN: "1new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1new",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name:   "AWS Role ARN",
			fields: baseFields,
			args: args{
				in: &Configuration{
					AWSRoleARN: "2new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1",
				AWSRoleARN:     "2new",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name:   "Okta Client ID",
			fields: baseFields,
			args: args{
				in: &Configuration{
					OktaClientID: "3new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3new",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name:   "Okta App ID",
			fields: baseFields,
			args: args{
				in: &Configuration{
					OktaAppID: "4new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "4new",
				OktaURL:        "5",
			},
		},
		{
			name:   "Okta URL",
			fields: baseFields,
			args: args{
				in: &Configuration{
					OktaURL: "5new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5new",
			},
		},
		{
			name:   "All AWS",
			fields: baseFields,
			args: args{
				in: &Configuration{
					AWSProviderARN: "1new",
					AWSRoleARN:     "2new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1new",
				AWSRoleARN:     "2new",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name:   "All Okta",
			fields: baseFields,
			args: args{
				in: &Configuration{
					OktaClientID: "3new",
					OktaAppID:    "4new",
					OktaURL:      "5new",
				},
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3new",
				OktaAppID:      "4new",
				OktaURL:        "5new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Configuration{
				AWSProviderARN: tt.fields.AWSProviderARN,
				AWSRoleARN:     tt.fields.AWSRoleARN,
				OktaClientID:   tt.fields.OktaClientID,
				OktaAppID:      tt.fields.OktaAppID,
				OktaURL:        tt.fields.OktaURL,
			}
			c.OverrideWith(tt.args.in)

			if !reflect.DeepEqual(c, tt.wantCfg) {
				fmt.Printf("Hm: %+v\n", c)
				t.Errorf("Configuration.OverrideWith() = %v, want %v", c, tt.wantCfg)
			}
		})
	}
}

func TestConfiguration_Validate(t *testing.T) {
	type fields struct {
		AWSProviderARN string
		AWSRoleARN     string
		OktaClientID   string
		OktaAppID      string
		OktaURL        string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name: "failure (missing AWSProviderARN)",
			fields: fields{
				AWSProviderARN: "",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
			wantErr: true,
		},
		{
			name: "failure (missing AWSRoleARN)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
			wantErr: true,
		},
		{
			name: "failure (missing OktaClientID)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
			wantErr: true,
		},
		{
			name: "failure (missing OktaAppID)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "",
				OktaURL:        "5",
			},
			wantErr: true,
		},
		{
			name: "failure (missing OktaURL)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Configuration{
				AWSProviderARN: tt.fields.AWSProviderARN,
				AWSRoleARN:     tt.fields.AWSRoleARN,
				OktaClientID:   tt.fields.OktaClientID,
				OktaAppID:      tt.fields.OktaAppID,
				OktaURL:        tt.fields.OktaURL,
			}

			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleNew() {
	cfg, err := New("my_profile", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("We have our configuration!", cfg)
}

func ExampleNew_with_override() {
	cfg, err := New("my_profile", "./path/to/config/config.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("We have our configuration!", cfg)
}

func ExampleConfiguration_OverrideWith() {
	cfg := Configuration{
		AWSProviderARN: "1",
		AWSRoleARN:     "2",
		OktaClientID:   "3",
		OktaAppID:      "4",
		OktaURL:        "5",
	}

	overrides := Configuration{
		AWSProviderARN: "1new",
		AWSRoleARN:     "2new",
	}

	cfg.OverrideWith(&overrides)

	fmt.Printf("Updated values: %+v\n", cfg)
}

func ExampleConfiguration_Validate() {
	cfg := Configuration{
		AWSProviderARN: "1",
		AWSRoleARN:     "2",
		OktaClientID:   "3",
		OktaURL:        "4",
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
}
