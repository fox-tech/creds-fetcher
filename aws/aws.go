// Package aws implements a Provider to interact with AWS
package aws

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/client"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/fsmanager"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/ini"
)

const (
	// stsURL represents the AWS STS URL to enchange SAML assertion
	// token for credentials
	stsURL = "https://sts.amazonaws.com/oauth2/v1/token"
	// TODO: How does this works for windows?
	credentialsDirectory = ".aws"
	credentialsFileName  = "credentials"
)

var (
	ErrBadRequest        = errors.New("invalid request to STS")
	ErrBadResponse       = errors.New("could not read response from STS")
	ErrFailedMarshal     = errors.New("encoding credentials failed")
	ErrFailedUnmarshal   = errors.New("decoding credentials failed")
	ErrFileHandlerFailed = errors.New("error handling file")
	ErrMissingProfile    = errors.New("profile required to create provider")
	ErrNotAuthorized     = errors.New("authentication failed")
	ErrUnknown           = errors.New("unexpected error ocurred")

	ioReadAll    = io.ReadAll
	iniMarshal   = ini.Marshal
	iniUnmarshal = ini.Unmarshal
)

// Profile indicates STS the principal and role to get credentials for
type Profile struct {
	Name         string
	RoleARN      string
	PrincipalARN string
}

// Provider exposes the methods to interact with AWS
type Provider struct {
	fs fileSystemManager
	httpClient

	Profile Profile
}

// IsEmpty verifies whether all fields of the profile are empty.
func (p Profile) IsEmpty() bool {
	return p.Name == "" || p.RoleARN == "" || p.PrincipalARN == ""
}

// httpClient defines the methods that the provider needs an http client to have
type httpClient interface {
	Get(r_url string, params map[string]string, body io.Reader) (*http.Response, error)
}

// fileSystemManager defines the methods that the provider needs a file system
// manager to have
type fileSystemManager interface {
	ReadFile(dir, filename string) ([]byte, error)
	WriteFile(name string, data []byte) error
}

// New returns a new provider with the given options.
// Returns error if no profile is set
func New(prf Profile, opts ...Option) (Provider, error) {
	p := Provider{
		fs:         fsmanager.NewDefault(),
		httpClient: client.NewDefault(),
		Profile:    prf,
	}

	if p.Profile.IsEmpty() {
		return p, ErrMissingProfile
	}

	for _, opt := range opts {
		opt(&p)
	}

	return p, nil
}

// GenerateCredentials requests AWS CLI credentials using a SAML assertion
// and saves them to a file
func (aws Provider) GenerateCredentials(saml string) error {
	// Exchange SAML for AWS Credentials
	cred, err := aws.getSTSCredentialsFromSAML(saml)
	if err != nil {
		return err
	}

	// Write credentials to file
	err = aws.updateCredentialsFile(cred)
	if err != nil {
		return err
	}

	return nil
}

// updateCredentialsFile reads exising credentials, adds or replaces the new credentials and saves them to file
func (p Provider) updateCredentialsFile(newCred credentials) error {
	log.Print("updating credentials file...")

	data, err := p.fs.ReadFile(credentialsDirectory, credentialsFileName)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFileHandlerFailed, err)
	}

	creds := map[string]credentials{}
	err = iniUnmarshal(data, creds)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedUnmarshal, err)
	}

	creds[p.Profile.Name] = newCred
	writeData, err := iniMarshal(creds)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedMarshal, err)
	}

	credentialsFilepath := filepath.Join(credentialsDirectory, credentialsFileName)
	if err = p.fs.WriteFile(credentialsFilepath, writeData); err != nil {
		return fmt.Errorf("%w: %v", ErrFileHandlerFailed, err)
	}

	log.Print("credentials saved to file")
	return nil
}
