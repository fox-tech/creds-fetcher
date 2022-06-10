package aws

import (
	"bytes"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/client"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/fsmanager"
)

type opts struct {
	p         Profile
	mckClient httpClient
	mckFs     fileSystemManager
}

func TestNew(t *testing.T) {
	type expect struct {
		p   Provider
		err error
	}

	type args struct {
		p    Profile
		opts []Option
	}

	emptyProfile := Profile{}

	prf := Profile{
		Name:         "test-profile",
		RoleARN:      "arn:aws:iam::ROLEARN",
		PrincipalARN: "arn:aws:iam::ProviderARN",
	}

	defaultClient := client.NewDefault()
	mckClient := client.MockHttpClient{}

	defaultFs := fsmanager.NewDefault()
	mckFs := fsmanager.NewMock(map[string][]byte{}, nil, nil)

	tests := []struct {
		name string
		args
		expect
	}{
		{
			name: "success: provider is created with default client and manager",
			args: args{p: prf},
			expect: expect{
				p: Provider{
					Profile:    prf,
					httpClient: defaultClient,
					fs:         defaultFs,
				},
				err: nil,
			},
		},
		{
			name: "success: provider is created mock client",
			args: args{
				p: prf,
				opts: []Option{
					setHTTPClient(mckClient),
				},
			},
			expect: expect{
				p: Provider{
					Profile:    prf,
					httpClient: mckClient,
					fs:         defaultFs,
				},
				err: nil,
			},
		},
		{
			name: "success: provider is created mock fs",
			args: args{
				p: prf,
				opts: []Option{
					setFileManager(mckFs),
				},
			},
			expect: expect{
				p: Provider{
					Profile:    prf,
					httpClient: defaultClient,
					fs:         mckFs,
				},
				err: nil,
			},
		},
		{
			name: "error: provider is created with empty profile",
			args: args{p: emptyProfile},
			expect: expect{
				p: Provider{
					Profile: Profile{},
				},
				err: ErrMissingProfile,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			np, err := New(tt.args.p, tt.args.opts...)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("New() expected error: %s, got: %s", tt.expect.err, err)
			}

			if np.Profile != tt.expect.p.Profile {
				t.Errorf("New() expected profile: %v, got %v", tt.expect.p.Profile, np.Profile)
			}

			// using reflection to validate the inner fields
			ect := reflect.TypeOf(tt.expect.p.httpClient)
			gct := reflect.TypeOf(np.httpClient)
			if ect != gct {
				t.Errorf("New() expected httpClient type: %v, got %v", ect, gct)
			}

			efst := reflect.TypeOf(tt.expect.p.fs)
			gfst := reflect.TypeOf(np.fs)
			if efst != gfst {
				t.Errorf("New() expected fsmanager type: %v, got %v", efst, gfst)
			}

		})
	}

}

func TestGetSTSCredentialsFromSAML(t *testing.T) {
	type expect struct {
		cred credentials
		err  error
	}

	saml := "reallylongsamlassertiong"
	prf := Profile{
		Name:         "test-profile",
		RoleARN:      "arn:aws:iam::ROLEARN",
		PrincipalARN: "arn:aws:iam::ProviderARN",
	}

	tests := []struct {
		name string
		arg  string
		opts
		expect
	}{
		{
			name: "request credentials, successful response: credentials are retrieved",
			arg:  saml,
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusOK,
					GetStatus:     "OK",
					GetBodyData:   []byte(successSTSResponse),
				},
			},
			expect: expect{
				cred: credentials{
					AccessKeyId:     "AWSACCESSKEYID",
					SecretAccessKey: "Super/Secret/AccessKey",
					SessionToken:    "reallylongandsecretsessiontoken",
				},
				err: nil,
			},
		},
		{
			name: "request credentials, error in getting credentials: error is returned",
			arg:  saml,
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetErr: errors.New("error while doing request"),
				},
			},
			expect: expect{
				err: ErrBadRequest,
			},
		},
		{
			name: "request credentials, error in reading body: error is returned",
			arg:  saml,
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusOK,
					GetStatus:     "OK",
				},
			},
			expect: expect{
				err: ErrBadResponse,
			},
		},
		{
			name: "request credentials, bad request: error is returned",
			arg:  saml,
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusBadRequest,
					GetStatus:     "Bad Request",
					GetBodyData:   []byte(errSTSResponse),
				},
			},
			expect: expect{
				err: ErrBadRequest,
			},
		},
		{
			name: "request credentials, unauthorized: error is returned",
			arg:  saml,
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusForbidden,
					GetStatus:     "Forbidden",
					GetBodyData:   []byte(errSTSResponse),
				},
			},
			expect: expect{
				err: ErrNotAuthorized,
			},
		},
		{
			name: "request credentials, unknown error: error is returned",
			arg:  saml,
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusInternalServerError,
					GetStatus:     "Internal Serverl Error",
				},
			},
			expect: expect{
				err: ErrUnknown,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := New(tt.opts.p,
				setHTTPClient(tt.opts.mckClient),
				setFileManager(tt.opts.mckFs),
			)

			cred, err := p.getSTSCredentialsFromSAML(tt.arg)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("getSTSCredentialsFromSAML() expected error: %s, got: %s", tt.expect.err, err)
			}

			if cred.AccessKeyId != tt.expect.cred.AccessKeyId ||
				cred.SecretAccessKey != tt.expect.cred.SecretAccessKey ||
				cred.SessionToken != tt.expect.cred.SessionToken {
				t.Errorf("getSTSCredentialsFromSAML() expected AccessKeyId: %v, got: %v", tt.expect.cred, cred)
			}
		})
	}
}

func TestUpdateCredentialsFile(t *testing.T) {
	type expect struct {
		data []byte
		err  error
	}

	cred := credentials{
		AccessKeyId:     "AWSACCESSKEYID",
		SecretAccessKey: "Super/Secret/AccessKey",
		SessionToken:    "reallylongandsecretsessiontoken",
	}

	prf := Profile{
		Name:         "test-profile",
		RoleARN:      "arn:aws:iam::ROLEARN",
		PrincipalARN: "arn:aws:iam::ProviderARN",
	}

	tests := []struct {
		name string
		arg  credentials
		expect
		opts
	}{
		{
			name: "new credentials: credentials are saved to file",
			arg:  cred,
			opts: opts{
				p:     prf,
				mckFs: fsmanager.NewMock(map[string][]byte{}, nil, nil),
			},
			expect: expect{
				data: []byte(newCredentialsFileContent),
				err:  nil,
			},
		},
		{
			name: "existing credentials: credentials are updated in file",
			arg:  cred,
			opts: opts{
				p: prf,
				mckFs: fsmanager.NewMock(map[string][]byte{
					".aws/credentials": []byte(credentialsFileContent),
				}, nil, nil),
			},
			expect: expect{
				data: []byte(newCredentialsFileContent),
				err:  nil,
			},
		},
		{
			name: "error reading file: error is returned",
			arg:  cred,
			opts: opts{
				p:     prf,
				mckFs: fsmanager.NewMock(map[string][]byte{}, errors.New("broken pipe"), nil),
			},
			expect: expect{
				data: []byte{},
				err:  ErrFileHandlerFailed,
			},
		},
		{
			name: "error writing file: error is returned",
			arg:  cred,
			opts: opts{
				p:     prf,
				mckFs: fsmanager.NewMock(map[string][]byte{}, nil, errors.New("permission denied")),
			},
			expect: expect{
				data: []byte{},
				err:  ErrFileHandlerFailed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := New(tt.opts.p,
				setHTTPClient(tt.opts.mckClient),
				setFileManager(tt.opts.mckFs),
			)

			err := p.updateCredentialsFile(tt.arg)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("updateCredentialsFile() expected error: %s, got: %s", tt.expect.err, err)
			}

			savedData, _ := tt.opts.mckFs.ReadFile(credentialsDirectory, credentialsFileName)
			if !bytes.Equal(savedData, tt.expect.data) {
				t.Errorf("updateCredentialsFile() expected file data: %s, got: %s", tt.expect.data, savedData)
			}
		})
	}
}

func TestGenerateCredentials(t *testing.T) {
	prf := Profile{
		Name:         "test-profile",
		RoleARN:      "arn:aws:iam::ROLEARN",
		PrincipalARN: "arn:aws:iam::ProviderARN",
	}

	type expect struct {
		data []byte
		err  error
	}
	tests := []struct {
		name string
		arg  string
		expect
		opts
	}{
		{
			name: "sucess: credentials are generated",
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusOK,
					GetStatus:     "OK",
					GetBodyData:   []byte(successSTSResponse),
				},
				mckFs: fsmanager.NewMock(map[string][]byte{}, nil, nil),
			},
			expect: expect{
				err:  nil,
				data: []byte(newCredentialsFileContent),
			},
		},
		{
			name: "error: error getting credentials from sts",
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusForbidden,
					GetStatus:     "Forbiddend",
					GetBodyData:   []byte(errSTSResponse),
				},
				mckFs: fsmanager.NewMock(map[string][]byte{}, nil, nil),
			},
			expect: expect{
				err:  ErrNotAuthorized,
				data: []byte{},
			},
		},
		{
			name: "error: error saving credentials to file",
			opts: opts{
				p: prf,
				mckClient: client.MockHttpClient{
					GetStatusCode: http.StatusOK,
					GetStatus:     "OK",
					GetBodyData:   []byte(successSTSResponse),
				},
				mckFs: fsmanager.NewMock(map[string][]byte{}, nil, errors.New("permission denied")),
			},
			expect: expect{
				err:  ErrFileHandlerFailed,
				data: []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := New(tt.opts.p,
				setHTTPClient(tt.opts.mckClient),
				setFileManager(tt.opts.mckFs),
			)

			err := p.GenerateCredentials(tt.arg)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("GenerateCredentials() expected error: %s, got: %s", tt.expect.err, err)
			}

			savedData, _ := tt.opts.mckFs.ReadFile(credentialsDirectory, credentialsFileName)
			if !bytes.Equal(savedData, tt.expect.data) {
				t.Errorf("GenerateCredentials() expected file data: %s, got: %s", tt.expect.data, savedData)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		p      Profile
		expect bool
	}{
		{
			name: "non-empty Profile",
			p: Profile{
				Name:         "test-profile",
				RoleARN:      "role-arn",
				PrincipalARN: "principal-arn",
			},
			expect: false,
		},
		{
			name: "empty Profile Name",
			p: Profile{
				RoleARN:      "role-arn",
				PrincipalARN: "principal-arn",
			},
			expect: true,
		},
		{
			name: "empty Profile RoleARN",
			p: Profile{
				Name:         "test-profile",
				PrincipalARN: "principal-arn",
			},
			expect: true,
		},
		{
			name: "empty Profile principalARN",
			p: Profile{
				Name:    "test-profile",
				RoleARN: "role-arn",
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.p.IsEmpty()
			if e != tt.expect {
				t.Errorf("IsEmpty() expected: %v, got %v", tt.expect, e)
			}
		})
	}
}
