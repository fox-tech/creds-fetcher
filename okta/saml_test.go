package okta

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testClientInternalGetSAML struct {
	res    []byte
	status int
	token  accessToken
}

func newServerClientInternalGetSAML(input testClientInternalGetSAML) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if r.RequestURI != fmt.Sprintf("/login/token/sso?token=%s", input.token.AccessToken) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<html><body>wrong URI, expected ` + r.RequestURI + `</body></html>`))
			return
		}

		w.WriteHeader(input.status)
		w.Write(input.res)
	}))
}

func TestClientInternalGetSAML(t *testing.T) {
	cases := []struct {
		name     string
		id       string
		sso      accessToken
		srvInput testClientInternalGetSAML
		expect   string
		err      error
	}{
		{
			name: "200: got correct HTML with SAML Response",
			sso:  accessToken{AccessToken: "testSSOToken"},
			srvInput: testClientInternalGetSAML{
				res:    samlData,
				status: http.StatusOK,
			},
			expect: "dGhpcyBpcyBhIHRlc3QgZm9yIGJhc2U2NA==",
		},
		{
			name: "err: 200 OK but no SAML",
			sso:  accessToken{AccessToken: "testSSOToken"},
			srvInput: testClientInternalGetSAML{
				res:    []byte(`<html><body>empty</body></html>`),
				status: http.StatusOK,
			},
			expect: "",
			err:    ErrNoSAMLResponseFound,
		},
		{
			name: "err: 400 and response with empty",
			sso:  accessToken{AccessToken: "testSSOToken"},
			srvInput: testClientInternalGetSAML{
				res:    []byte(`empty`),
				status: http.StatusBadRequest,
			},
			expect: "",
			err:    ErrSAMLRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// pre-setup
			tt.srvInput.token = tt.sso
			srv := newServerClientInternalGetSAML(tt.srvInput)
			defer srv.Close()
			c := Client{uri: srv.URL}

			// actual test
			saml, err := c.getSAML(tt.sso)
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, received: %v", tt.err, err)
			}

			if err != nil {
				return // we got here because we expected an error and it matched
			}

			if saml != tt.expect {
				t.Errorf("expected %v, received %v", tt.expect, saml)
			}
		})
	}
}

var samlData = []byte(`<html><body><div><div><div><div>
<form method="post" action="https://sp.example.com/SAML2/SSO/POST" ...>
<input type="hidden" name="SAMLResponse" value="dGhpcyBpcyBhIHRlc3QgZm9yIGJhc2U2NA==" />
<input type="hidden" name="RelayState" value="token" />
<input type="submit" value="Submit" />
</div></div></div></div></form></body></html>`)

func TestInternalExtractSAML(t *testing.T) {
	cases := []struct {
		name   string
		expect string
		html   []byte
		err    error
	}{
		{
			name:   "good response",
			html:   samlData,
			expect: "dGhpcyBpcyBhIHRlc3QgZm9yIGJhc2U2NA==",
		},
		{
			name: "no tag",
			html: []byte(`<html><body></body></html>`),
			err:  ErrNoSAMLResponseFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			saml, err := extractSAML(tt.html)
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, received: %v", tt.err, err)
			}

			if err != nil {
				return // we got here because we expected an error and it matched
			}

			if saml != tt.expect {
				t.Errorf("expected %v, received %v", tt.expect, saml)
			}
		})
	}
}
