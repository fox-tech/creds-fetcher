package aws

const (
	defatulProfile = "default"
	stsURL         = "https://sts.amazonaws.com/oauth2/v1/token"
	// TODO: How does this works for windows?
	credentialsDirectory = ".aws"
	credentialsFileName  = "credentials"
)

type Profile struct {
	Name         string
	RoleARN      string
	PrincipalARN string
}

type Provider struct {
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
