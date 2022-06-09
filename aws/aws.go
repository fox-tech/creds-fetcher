package aws

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/client"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/fsmanager"
	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/ini"
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
)

// New returns a new provider with the given options.
// Returns error if no profile is set
func New(opts ...Option) (Provider, error) {
	aws := Provider{}

	for _, opt := range opts {
		opt(&aws)
	}

	if aws.Profile.IsEmpty() {
		return aws, ErrMissingProfile
	}

	if aws.fs == nil {
		aws.fs = fsmanager.NewDefault()
	}

	if aws.httpClient == nil {
		aws.httpClient = client.NewDefault()
	}

	return aws, nil
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

// getSTSCredentialsFromSAML uses provided saml string to requests AWS CLI
// credentials using STS.
func (aws Provider) getSTSCredentialsFromSAML(saml string) (credentials, error) {
	log.Print("getting STS credentials...")

	params := map[string]string{
		"Version":       "2011-06-15",
		"Action":        "AssumeRoleWithSAML",
		"RoleArn":       aws.Profile.RoleARN,
		"PrincipalArn":  aws.Profile.PrincipalARN,
		"SAMLAssertion": saml,
	}

	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}

	req, err := http.NewRequest(http.MethodGet, stsURL, nil)
	if err != nil {
		return credentials{}, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	req.URL.RawQuery = q.Encode()
	resp, err := aws.httpClient.Do(req)
	if err != nil {
		return credentials{}, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return credentials{}, fmt.Errorf("%w: %v", ErrBadResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		errResponse := assumeRoleWiwthSAMLError{}

		// ignoring unmarshall error to continue in case response does not have body
		xml.Unmarshal(respBody, &errResponse)

		switch resp.StatusCode {
		case http.StatusBadRequest:
			return credentials{}, fmt.Errorf("%w: status: %scode: %s message: %s", ErrBadRequest, resp.Status, errResponse.Error.Code, errResponse.Error.Message)
		case http.StatusForbidden:
			return credentials{}, fmt.Errorf("%w: status: %s, code: %s message: %s", ErrNotAuthorized, resp.Status, errResponse.Error.Code, errResponse.Error.Message)
		default:
			return credentials{}, fmt.Errorf("%w: status: %s, code: %s message: %s", ErrUnknown, resp.Status, errResponse.Error.Code, errResponse.Error.Message)
		}
	}

	stsResp := assumeRoleWithSAMLResponse{}
	if err := xml.Unmarshal(respBody, &stsResp); err != nil {
		return credentials{}, fmt.Errorf("%w: could not unmarshall response: %v", ErrBadResponse, err)
	}

	log.Print("STS credentials retrieved")

	return stsResp.AssumeRoleResult.Credentials, nil
}

// updateCredentialsFile reads exising credentials, adds or replaces the new credentials and saves them to file
func (p Provider) updateCredentialsFile(newCred credentials) error {
	log.Print("updating credentials file...")

	data, err := p.fs.ReadFile(credentialsDirectory, credentialsFileName)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFileHandlerFailed, err)
	}

	creds := map[string]credentials{}
	err = ini.Unmarshal(data, creds)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedUnmarshal, err)
	}

	creds[p.Profile.Name] = newCred
	writeData, err := ini.Marshal(creds)
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
