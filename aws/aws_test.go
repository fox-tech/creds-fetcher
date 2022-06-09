package aws

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/client"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/fsmanager"
)

type opts struct {
	p         Profile
	mckClient httpClient
	mckFs     fileSystemManager
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
				p:         prf,
				mckClient: client.NewMock(http.StatusOK, "OK", []byte(successSTSResponse), nil),
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
				p:         prf,
				mckClient: client.NewMock(0, "OK", []byte{}, errors.New("error while doing request")),
			},
			expect: expect{
				err: ErrBadRequest,
			},
		},
		{
			name: "request credentials, bad request: error is returned",
			arg:  saml,
			opts: opts{
				p:         prf,
				mckClient: client.NewMock(http.StatusBadRequest, "BadRequest", []byte(errSTSResponse), nil),
			},
			expect: expect{
				err: ErrBadRequest,
			},
		},
		{
			name: "request credentials, unauthorized: error is returned",
			arg:  saml,
			opts: opts{
				p:         prf,
				mckClient: client.NewMock(http.StatusForbidden, "Unauthorized", []byte(errSTSResponse), nil),
			},
			expect: expect{
				err: ErrNotAuthorized,
			},
		},
		{
			name: "request credentials, unknown error: error is returned",
			arg:  saml,
			opts: opts{
				p:         prf,
				mckClient: client.NewMock(http.StatusInternalServerError, "Internal Server Error", []byte{}, nil),
			},
			expect: expect{
				err: ErrUnknown,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(SetProfile(tt.opts.p),
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
			p := New(SetProfile(tt.opts.p),
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
				p:         prf,
				mckClient: client.NewMock(http.StatusOK, "OK", []byte(successSTSResponse), nil),
				mckFs:     fsmanager.NewMock(map[string][]byte{}, nil, nil),
			},
			expect: expect{
				err:  nil,
				data: []byte(newCredentialsFileContent),
			},
		},
		{
			name: "error: error getting credentials from sts",
			opts: opts{
				p:         prf,
				mckClient: client.NewMock(http.StatusForbidden, "Unathorized", []byte(errSTSResponse), nil),
				mckFs:     fsmanager.NewMock(map[string][]byte{}, nil, nil),
			},
			expect: expect{
				err:  ErrNotAuthorized,
				data: []byte{},
			},
		},
		{
			name: "error: error saving credentials to file",
			opts: opts{
				p:         prf,
				mckClient: client.NewMock(http.StatusOK, "OK", []byte(successSTSResponse), nil),
				mckFs:     fsmanager.NewMock(map[string][]byte{}, nil, errors.New("permission denied")),
			},
			expect: expect{
				err:  ErrFileHandlerFailed,
				data: []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(SetProfile(tt.opts.p),
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
