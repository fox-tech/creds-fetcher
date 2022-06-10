package configuration

import (
	"os"
	"reflect"
	"testing"
)

const (
	exampleJSON = `
{
	"aws_provider_arn" : "1",
	"aws_role_arn" : "2",
	"okta_client_id" : "3",
	"okta_url" : "4"				
}
`

	exampleJSONInvalid = `
{
	"aws_role_arn" : "2",
	"okta_client_id" : "3",
	"okta_url" : "4"				
}
`
	exampleJSONArray = `["hello world", "foo", "bar", "baz"]`

	exampleTOML = `
aws_provider_arn = "1"
aws_role_arn = "2"
okta_client_id = "3"
okta_url = "4"
`
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		prep func() (toRemove *os.File, err error)

		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success",
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSON)
				os.Stdin = toRemove
				return
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaURL:        "4",
			},
		},
		{
			name: "failure (invalid configuration)",
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSONInvalid)
				os.Stdin = toRemove
				return
			},
			wantErr: true,
		},
		{
			name: "failure (closed file)",
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

			gotCfg, err := New()
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

func TestConfiguration_Validate(t *testing.T) {
	type fields struct {
		AWSProviderARN string
		AWSRoleARN     string
		OktaClientID   string
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
				OktaURL:        "4",
			},
		},
		{
			name: "failure (missing AWSProviderARN)",
			fields: fields{
				AWSProviderARN: "",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
				OktaURL:        "4",
			},
			wantErr: true,
		},
		{
			name: "failure (missing AWSRoleARN)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "",
				OktaClientID:   "3",
				OktaURL:        "4",
			},
			wantErr: true,
		},
		{
			name: "failure (missing OktaClientID)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "",
				OktaURL:        "4",
			},
			wantErr: true,
		},
		{
			name: "failure (missing OktaURL)",
			fields: fields{
				AWSProviderARN: "1",
				AWSRoleARN:     "2",
				OktaClientID:   "3",
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
				OktaURL:        tt.fields.OktaURL,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
