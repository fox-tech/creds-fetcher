package aws

import (
	"net/http"
)

const (
	defatulProfile = "default"
	stsURL         = "https://sts.amazonaws.com/oauth2/v1/token"
	// TODO: How does this works for windows?
	credentialsDirectory = ".aws"
	credentialsFileName  = "credentials"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type fileSystemManager interface {
	ReadFile(dir, filename string) ([]byte, error)
	WriteFile(name string, data []byte) error
}

type Profile struct {
	Name         string
	RoleARN      string
	PrincipalARN string
}

type Provider struct {
	fs fileSystemManager
	httpClient

	Profile Profile
}

type assumeRoleWithSAMLResponse struct {
	AssumeRoleResult assumeRoleResult `xml:"AssumeRoleWithSAMLResult"`
}

type assumeRoleResult struct {
	Credentials credentials `xml:"Credentials"`
}

type credentials struct {
	AccessKeyId     string `xml:"AccessKeyId" ini:"aws_access_key_id"`
	SecretAccessKey string `xml:"SecretAccessKey" ini:"aws_secret_access_key"`
	SessionToken    string `xml:"SessionToken" ini:"aws_session_token"`
}

type assumeRoleWiwthSAMLError struct {
	Error stsError `xml:"Error"`
}

type stsError struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

func (p Profile) IsEmpty() bool {
	return p.Name == "" || p.RoleARN == "" || p.PrincipalARN == ""
}
