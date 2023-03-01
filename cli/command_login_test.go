package cli

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fox-tech/creds-fetcher/aws"
)

type testServerInput struct {
	code     int
	response []byte
}

func newTestServer(input map[string]testServerInput) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var code int
		var response []byte
		switch r.RequestURI {
		case "/oauth2/v1/device/authorize":
			code = input["authorize"].code
			response = input["authorize"].response
		case "/oauth2/v1/token":
			code = input["token"].code
			response = input["token"].response
		case "/login/token/sso?token=accesstoken":
			code = input["sso"].code
			response = input["sso"].response
		// default is for STS response since it doesn't contain any uri
		default:
			code = input["sts"].code
			response = input["sts"].response
		}
		w.WriteHeader(code)
		w.Write(response)
	}))
}

func createConfigFile(config string) *os.File {
	file, err := os.Create("./test-config.toml")
	if err != nil {
		log.Fatal("could not create temporary file: %w", err)
	}
	_, err = file.Write([]byte(config))
	if err != nil {
		log.Fatal("could not create temporary file: %w", err)
	}
	if err = file.Close(); err != nil {
		log.Fatal("could not create temporary file: %w", err)
	}
	return file
}

func removeConfigFile(f *os.File) {
	os.Remove(f.Name())
}

func Test_login(t *testing.T) {
	config := `
	[test]
	aws_provider_arn = "arn:aws:iam::provider"
	aws_role_arn  = "arn:aws:iam::role"
	okta_client_id = "123"
	okta_app_id = "234"
	`

	type args struct {
		flags     FlagMap
		responses map[string]testServerInput
	}

	tests := []struct {
		name   string
		expect error
		args
	}{
		{
			name: "successful login with named profile",
			args: args{
				flags: FlagMap{
					FlagProfile: Flag{Name: "profile", Value: "test"},
					FlagConfig:  Flag{Name: "config", Value: "test-config.toml"},
				},
				responses: map[string]testServerInput{
					"authorize": {
						code:     http.StatusOK,
						response: []byte(`{"device_code": "9b", "user_code": "D", "verification_uri": "activate", "verification_uri_complete": "oktap.com/activate?user_code=D", "expires_in":10,"interval": 5}`),
					},
					"token": {
						code:     http.StatusOK,
						response: []byte(`{"access_token": "accesstoken"}`),
					},
					"sso": {
						code:     http.StatusOK,
						response: []byte(`<div><input name="SAMLResponse" value="token"/></div>`),
					},
					"sts": {
						code:     http.StatusOK,
						response: []byte(aws.SuccessSTSResponse),
					},
				},
			},
		},
		{
			name: "error: profile flag not found",
			args: args{
				flags: FlagMap{
					FlagConfig: Flag{Name: "config", Value: "test-config.toml"},
				},
			},
			expect: ErrNotFound,
		},
		{
			name: "error: config flag not found",
			args: args{
				flags: FlagMap{
					FlagProfile: Flag{Name: "profile", Value: "test"},
				},
			},
			expect: ErrNotFound,
		},
		{
			name: "error: configuration not found",
			args: args{
				flags: FlagMap{
					FlagProfile: Flag{Name: "profile", Value: ""},
					FlagConfig:  Flag{Name: "config", Value: "test-config.toml"},
				},
			},
			expect: ErrNoConfig,
		},
		{
			name: "error: okta preauthorize error",
			args: args{
				flags: FlagMap{
					FlagProfile: Flag{Name: "profile", Value: "test"},
					FlagConfig:  Flag{Name: "config", Value: "test-config.toml"},
				},
				responses: map[string]testServerInput{
					"authorize": {
						code:     http.StatusOK,
						response: []byte(`{`),
					},
				},
			},
			expect: ErrAuthenticationFailed,
		},
		{
			name: "error: okta authorize error",
			args: args{
				flags: FlagMap{
					FlagProfile: Flag{Name: "profile", Value: "test"},
					FlagConfig:  Flag{Name: "config", Value: "test-config.toml"},
				},
				responses: map[string]testServerInput{
					"authorize": {
						code:     http.StatusOK,
						response: []byte(`{"device_code": "9b", "user_code": "D", "verification_uri": "activate", "verification_uri_complete": "oktap.com/activate?user_code=D", "expires_in":10,"interval": 5}`),
					},
					"token": {
						code:     http.StatusOK,
						response: []byte(`{"access_token": "accesstoken"}`),
					},
					"sso": {
						code:     http.StatusOK,
						response: []byte(`<div><input name="SAMLResponse" value="token"/></div>`),
					},
					"sts": {
						code:     http.StatusForbidden,
						response: []byte(""),
					},
				},
			},
			expect: ErrAuthenticationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := newTestServer(tt.args.responses)
			defer s.Close()

			f := createConfigFile(fmt.Sprintf("%sokta_url = \"%s\"\n", config, s.URL))
			defer removeConfigFile(f)

			prevURL := aws.STSURL
			aws.STSURL = s.URL
			defer func() { aws.STSURL = prevURL }()

			err := login(tt.args.flags)

			if !errors.Is(err, tt.expect) {
				t.Errorf("login() expected error: %v, got %v", tt.expect, err)
			}
		})
	}
}
