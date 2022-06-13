package okta

type mockProvider struct {
	Provider
}

func (m mockProvider) GenerateCredentials(assert string) error {
	return nil
}
