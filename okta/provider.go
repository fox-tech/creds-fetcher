package okta

// Provider is the interface that wraps the GenerateCredentials method.
//
// GenerateCredentials implements the logic on the provider-side to convert the
// passed SAML assertion into provider-ready credentials. It must save the
// generated credentials in a way the provider's client will be able to use it.
// Must return nil error un success.
type Provider interface {
	GenerateCredentials(assertion string) error
}
