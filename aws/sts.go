package aws

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
)

// assumeRoleWithSAMLResponse represents part of the STS response to an
// AssumeRoleWithSAML request when it was successful
type assumeRoleWithSAMLResponse struct {
	AssumeRoleResult assumeRoleResult `xml:"AssumeRoleWithSAMLResult"`
}

// assumeRoleResult contains the credentials returned from a successful STS
// AssumeRoleWithSAML request
type assumeRoleResult struct {
	Credentials credentials `xml:"Credentials"`
}

// credentials represents the login values returned by STS
type credentials struct {
	AccessKeyId     string `xml:"AccessKeyId" ini:"aws_access_key_id"`
	SecretAccessKey string `xml:"SecretAccessKey" ini:"aws_secret_access_key"`
	SessionToken    string `xml:"SessionToken" ini:"aws_session_token"`
}

// assumeRoleWithSAMLError represents part of the STS response to an
// AssumeRoleWithSAML request when it failed
type assumeRoleWithSAMLError struct {
	Error stsError `xml:"Error"`
}

// stsError represents the error message returned from STS when the
// request failed.
type stsError struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

// getSTSCredentialsFromSAML uses provided saml string to requests AWS CLI
// credentials using STS.
func (p Provider) getSTSCredentialsFromSAML(saml string) (credentials, error) {
	log.Print("getting STS credentials...")

	params := map[string]string{
		"Version":       "2011-06-15",
		"Action":        "AssumeRoleWithSAML",
		"RoleArn":       p.Profile.RoleARN,
		"PrincipalArn":  p.Profile.PrincipalARN,
		"SAMLAssertion": saml,
	}

	resp, err := p.httpClient.Get(stsURL, params, nil)
	defer resp.Body.Close()
	if err != nil {
		return credentials{}, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	respBody, err := ioReadAll(resp.Body)
	if err != nil {
		return credentials{}, fmt.Errorf("%w: %v", ErrBadResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		errResponse := assumeRoleWithSAMLError{}

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
