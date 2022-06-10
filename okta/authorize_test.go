package okta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testClientPreAuthorize struct {
	res      []byte
	clientID string
}

func newServerClientPreAuthorize(input testClientPreAuthorize, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		expectedURI := "/oauth2/v1/device/authorize"
		if r.RequestURI != expectedURI {
			w.WriteHeader(http.StatusBadRequest)
			errString := fmt.Sprintf(`{"error":"badURL","error_description":"requested %s and should be %s"}`, r.RequestURI, expectedURI)
			w.Write([]byte(errString))
			return
		}

		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		expectedPayload := fmt.Sprintf("client_id=%s&scope=openid+okta.apps.sso", input.clientID)
		if string(body) != expectedPayload {
			w.WriteHeader(http.StatusNotAcceptable)
			errString := fmt.Sprintf(`{"error":"badPayload","error_description":"received %s and should be %s"}`, string(body), expectedPayload)
			w.Write([]byte(errString))
			return
		}

		w.WriteHeader(status)
		if status == http.StatusNotAcceptable {
			w.Write([]byte(`{{`))
		}

		w.Write(input.res)
	}))
}

func TestClientPreAuthorize(t *testing.T) {
	cases := []struct {
		name    string
		expect  Device
		errBody errorResponse
		id      string
		status  int
		err     error
	}{
		{
			name:   "200: good body, good request",
			id:     "testid",
			status: http.StatusOK,
			expect: Device{
				DeviceCode:              "iiuiudw99ajksjkdaw",
				UserCode:                "HJJDS87SS",
				VerificationURI:         "https://randomrandom.okta.com/activate",
				VerificationURIComplete: "https://randomrandom.okta.com/activate?user_code=HJJDS87SS",
				ExpiresIn:               600,
				Interval:                5,
			},
		},
		{
			name:   "400: bad request",
			status: http.StatusBadRequest,
		},
		{
			name:   "force json decoding error",
			status: http.StatusNotAcceptable,
			err:    ErrPreAuthorizeJSONDecode,
		},
		{
			name:   "503: forcing server error",
			status: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// pre-setup
			buf := new(bytes.Buffer)
			if tt.err == nil {
				err := json.NewEncoder(buf).Encode(tt.expect)
				if err != nil {
					t.Fatalf("decoding expected result: %v", err)
				}
			} else {
				err := json.NewEncoder(buf).Encode(tt.errBody)
				if err != nil {
					t.Fatalf("decoding expected error: %v", err)
				}
			}

			input := testClientPreAuthorize{
				res:      buf.Bytes(),
				clientID: tt.id,
			}

			srv := newServerClientPreAuthorize(input, tt.status)
			defer srv.Close()
			c := Client{uri: srv.URL, id: tt.id}

			// actual test
			res, err := c.PreAuthorize()
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, received: %v", tt.err, err)
			}

			if err != nil {
				return // we got here because we expected an error and it matched
			}

			if res.DeviceCode != tt.expect.DeviceCode {
				t.Errorf("expected %v, received: %v", tt.expect, res)
			}
		})
	}
}

type testClientAuthorize struct {
	clientID     string
	pollResponse accessToken
	pollStatus   int
	ssoResponse  accessToken
	ssoStatus    int
	samlResponse []byte
	samlStatus   int
}

func newServerClientAuthorize(input testClientAuthorize) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/oauth2/v1/token": // token poll or sso token
			w.Header().Set("Content-Type", "application/json")
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"bad_payload","error_description":"` + err.Error() + `}`))
				return
			}

			buf := new(bytes.Buffer)
			if strings.Contains(string(body), "device_code") { // token poll
				err := json.NewEncoder(buf).Encode(input.pollResponse)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"bad_encoding","error_description":"` + err.Error() + `}`))
				}
				w.WriteHeader(input.pollStatus)
				w.Write(buf.Bytes())
				return
			} else { // sso token
				err := json.NewEncoder(buf).Encode(input.ssoResponse)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"bad_encoding","error_description":"` + err.Error() + `}`))
				}
				w.WriteHeader(input.ssoStatus)
				w.Write(buf.Bytes())
				return
			}

		default:
			w.Header().Set("Content-Type", "text/html")
			if !strings.Contains(r.RequestURI, "/login/token/sso?token=") {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(input.samlStatus)
			w.Write(input.samlResponse)
		}
	}))
}

func TestClientAuthorize(t *testing.T) {
	defaultDevice := Device{
		DeviceCode:              "b33fid-d3v1c3c0d3",
		UserCode:                "JU9992JU",
		VerificationURI:         "https://randomrandom.okta.com/activate",
		VerificationURIComplete: "https://randomrandom.okta.com/activate?user_code=JU9992JU",
		ExpiresIn:               5,
		Interval:                1,
	}

	cases := []struct {
		name     string
		id       string
		err      error
		device   Device
		srvInput testClientAuthorize
	}{
		{
			name:   "200: all good",
			device: defaultDevice,
			id:     "testid",
			srvInput: testClientAuthorize{
				clientID:     "testid",
				pollResponse: accessToken{AccessToken: "random_accesstoken", IDToken: "random_idtoken"},
				pollStatus:   http.StatusOK,
				ssoResponse:  accessToken{AccessToken: "random_ssotoken"},
				ssoStatus:    http.StatusOK,
				samlResponse: samlData,
				samlStatus:   http.StatusOK,
			},
		},
		{
			name:   "500 on pollResponse",
			device: defaultDevice,
			id:     "testid",
			srvInput: testClientAuthorize{
				clientID:     "testid",
				pollResponse: accessToken{AccessToken: "random_accesstoken", IDToken: "random_idtoken"},
				pollStatus:   http.StatusInternalServerError,
				ssoResponse:  accessToken{AccessToken: "random_ssotoken"},
				ssoStatus:    http.StatusOK,
				samlResponse: samlData,
				samlStatus:   http.StatusOK,
			},
			err: ErrAccessTokenRequest,
		},
		{
			name:   "500 on ssoResponse",
			device: defaultDevice,
			id:     "testid",
			srvInput: testClientAuthorize{
				clientID:     "testid",
				pollResponse: accessToken{AccessToken: "random_accesstoken", IDToken: "random_idtoken"},
				pollStatus:   http.StatusOK,
				ssoResponse:  accessToken{AccessToken: "random_ssotoken"},
				ssoStatus:    http.StatusInternalServerError,
				samlResponse: samlData,
				samlStatus:   http.StatusOK,
			},
			err: ErrSSORequest,
		},
		{
			name:   "500 on getSAML",
			device: defaultDevice,
			id:     "testid",
			srvInput: testClientAuthorize{
				clientID:     "testid",
				pollResponse: accessToken{AccessToken: "random_accesstoken", IDToken: "random_idtoken"},
				pollStatus:   http.StatusOK,
				ssoResponse:  accessToken{AccessToken: "random_ssotoken"},
				ssoStatus:    http.StatusOK,
				samlResponse: samlData,
				samlStatus:   http.StatusInternalServerError,
			},
			err: ErrSAMLRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// pre-setup
			srv := newServerClientAuthorize(tt.srvInput)
			defer srv.Close()

			c, err := New(tt.id, srv.URL, mockProvider{})
			if err != nil {
				t.Fatalf("unexpected error initializing Client: %v", err)
			}

			// actual test
			err = c.Authorize(tt.device)
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, received: %v", tt.err, err)
			}
		})
	}

}
