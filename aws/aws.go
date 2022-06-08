package aws

import (
	"encoding/xml"
	"errors"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/foxbroadcasting/fox-okta-oie-gimme-aws-creds/ini"
)

var (
	ErrBadRequest        = errors.New("Invalid request to STS")
	ErrNotAuthorized     = errors.New("Authentication failed")
	ErrUnknown           = errors.New("Unexpected error ocurred")
	ErrBadResponse       = errors.New("Could not read response from STS")
	ErrCouldNotReadFile  = errors.New("Read credentials from file failed")
	ErrCouldNotWriteFile = errors.New("Write credentials to file failed")
)

func New(opts ...Option) Provider {
	aws := Provider{}

	for _, opt := range opts {
		opt(&aws)
	}

	return aws
}

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

func (aws Provider) getSTSCredentialsFromSAML(saml string) (credentials, error) {
	log.Print("getting STS credentials...")

	params := map[string]string{
		"Version":       "2011-06-15",
		"Action":        "AssumeRoleWithSAML",
		"RoleArn":       aws.Profile.RoleARN,
		"PrincipalArn":  aws.Profile.PrincipalARN,
		"SAMLAssertion": saml,
	}

	client := &http.Client{}
	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}

	req, err := http.NewRequest(http.MethodGet, stsURL, nil)
	if err != nil {
		log.Fatalf("error creating request to STS: %v", err)
		return credentials{}, ErrBadRequest
	}

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error in response from STS: %v", err)
		return credentials{}, ErrBadRequest
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return credentials{}, ErrBadResponse
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error in response from STS: %s", resp.Status)
		// TODO: Add parsing of error message in body
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return credentials{}, ErrBadRequest
		case http.StatusForbidden:
			return credentials{}, ErrNotAuthorized
		default:
			return credentials{}, ErrUnknown
		}
	}

	stsResp := assumeRoleWithSAMLResponse{}
	if err = xml.Unmarshal(respBody, &stsResp); err != nil {
		log.Fatal("could not unmarshal STS response")
		return credentials{}, ErrBadResponse
	}

	log.Print("STS credentials retrieved")

	return stsResp.AssumeRoleResult.Credentials, nil
}

func (aws Provider) updateCredentialsFile(newCred credentials) error {
	log.Print("updating credentials file...")

	data, err := readCredentialsFile(credentialsDirectory, credentialsFileName)
	if err != nil {
		log.Fatalf("unable to open credentials file: %v", err)
		return ErrCouldNotReadFile
	}

	creds := map[string]credentials{}
	err = ini.Unmarshal(data, creds)
	if err != nil {
		log.Fatalf("error while reading credentials file: %v", err)
		return ErrCouldNotReadFile
	}

	creds[aws.Profile.Name] = newCred
	writeData, err := ini.Marshal(creds)
	if err != nil {
		log.Fatalf("Unable to write credentials file: %v", err)
		return ErrCouldNotWriteFile
	}

	credentialsFilepath := filepath.Join(credentialsDirectory, credentialsFileName)
	if err = os.WriteFile(credentialsFilepath, writeData, fs.FileMode(os.O_RDWR)); err != nil {
		log.Fatalf("Unable to write credentials file: %v", err)
		return ErrCouldNotWriteFile
	}

	log.Print("credentials saved to file")
	return nil
}

func readCredentialsFile(dir, filename string) (data []byte, err error) {
	fp := filepath.Join(dir, filename)

	// Switch to home directory
	hd, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to open credentials file: %v", err)
	}

	if err = os.Chdir(hd); err != nil {
		log.Fatalf("unable to open credentials file: %v", err)
	}

	data, err = ioutil.ReadFile(fp)
	if err == nil || (err != nil && !os.IsNotExist(err)) {
		return
	}

	_, err = os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}

		log.Print("credentials directory not found, creating...")
		if err = os.Mkdir(dir, 0766); err != nil {
			return
		}
		log.Print("credentials directory created")
	}

	log.Print("credentials file not found, creating...")
	_, err = os.Create(fp)
	if err != nil {
		return
	}
	log.Printf("credentials file created: %s", fp)

	return
}
