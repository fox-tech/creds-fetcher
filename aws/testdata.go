package aws

const successSTSResponse = `<AssumeRoleWithSAMLResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<AssumeRoleWithSAMLResult>
    <Audience>https://signin.aws.amazon.com/saml</Audience>
    <AssumedRoleUser>
      <AssumedRoleId>AROARORTY3BBGGVCOV4OP:maria.ramirez@wizeline.com</AssumedRoleId>
      <Arn>arn:aws:sts::099966572610:assumed-role/okta-oie-ReadOnly/maria.ramirez@wizeline.com</Arn>
    </AssumedRoleUser>
    <Credentials>
      <AccessKeyId>AWSACCESSKEYID</AccessKeyId>
      <SecretAccessKey>Super/Secret/AccessKey</SecretAccessKey>
      <SessionToken>reallylongandsecretsessiontoken</SessionToken>
      <Expiration>2022-06-07T22:54:14Z</Expiration>
    </Credentials>
    <Subject>maria.ramirez@wizeline.com</Subject>
    <NameQualifier>ntHNa4L+06nVQFiE3ekTsihyVQY=</NameQualifier>
    <SubjectType>unspecified</SubjectType>
    <Issuer>http://www.okta.com/exk3kzb4ja8B9PHgs1d7</Issuer>
  </AssumeRoleWithSAMLResult>
  <ResponseMetadata>
    <RequestId>27519a21-569b-401b-92e1-a88fe12661e8</RequestId>
  </ResponseMetadata>
</AssumeRoleWithSAMLResponse>
`

const errSTSResponse = `
<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>ExpiredTokenException</Code>
  <Message>Token must be redeemed within 5 minutes of issuance</Message>
</Error>
<RequestId>51de7dff-3803-47db-b8a7-4430a295e699</RequestId>
</ErrorResponse>
`

const credentialsFileContent = "[test-profile]\naws_access_key_id = AWSACCESSKEYID\naws_secret_access_key = Super/Secret/AccessKey\naws_session_token = oldreallylongandreallysecrettoken\n\n"
const newCredentialsFileContent = "[test-profile]\naws_access_key_id = AWSACCESSKEYID\naws_secret_access_key = Super/Secret/AccessKey\naws_session_token = reallylongandsecretsessiontoken\n\n"
