package okta

import "testing"

func TestNew(t *testing.T) {
	cases := []struct {
		name     string
		id       string
		uri      string
		provider Provider
		opts     []Option
		err      error
	}{
		{
			name: "no values",
			err:  ErrMissingClientConfig,
		},
		{
			name:     "no id",
			uri:      "http://random.com",
			provider: mockProvider{},
			err:      ErrMissingClientConfig,
		},
		{
			name:     "no uri",
			id:       "b33fid",
			provider: mockProvider{},
			err:      ErrMissingClientConfig,
		},
		{
			name: "no provider",
			id:   "b33fid",
			uri:  "http://random.com",
			err:  ErrNoProvider,
		},
		{
			name:     "optional appID",
			id:       "b33fid",
			uri:      "http://random.com",
			provider: mockProvider{},
			opts:     []Option{SetAppID("anAppID")},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.id, tt.uri, tt.provider, tt.opts...)
			if err != nil && err != tt.err {
				t.Fatalf("expected error %v and received %v", tt.err, err)
			}
		})
	}
}
