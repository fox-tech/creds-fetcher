package okta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testClientInternalAccessTokenPoll struct {
	res        []byte
	clientID   string
	deviceCode string
	intervals  int
	status     int
	srv        *httptest.Server
	count      int
}

func (s *testClientInternalAccessTokenPoll) newServerClientInternalAccessTokenPoll() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		expectedURI := "/oauth2/v1/token"
		if r.RequestURI != expectedURI {
			w.WriteHeader(http.StatusBadRequest)
			errString := fmt.Sprintf(`{"error":"badURL","error_description":"requested %s and should be %s"}`, r.RequestURI, expectedURI)
			w.Write([]byte(errString))
			return
		}

		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		expectedPayload := url.Values{
			"client_id":   []string{s.clientID},
			"device_code": []string{s.deviceCode},
			"grant_type":  []string{"urn:ietf:params:oauth:grant-type:device_code"},
		}.Encode()

		if string(body) != expectedPayload {
			w.WriteHeader(http.StatusNotAcceptable)
			errString := fmt.Sprintf(`{"error":"badPayload","error_description":"received %s and should be %s"}`, string(body), expectedPayload)
			w.Write([]byte(errString))
			return
		}

		if s.count < s.intervals {
			w.WriteHeader(http.StatusBadRequest)
			s.count++
			return
		}

		w.WriteHeader(s.status)
		w.Write(s.res)
	}))
}

func TestClientInternalAccessTokenPoll(t *testing.T) {
	cases := []struct {
		name      string
		id        string
		device    Device
		expect    accessToken
		err       error
		errBody   string
		status    int
		intervals int
	}{
		{
			name: "200: good request in 3 intervals",
			device: Device{
				DeviceCode:              "b33fid",
				UserCode:                "FJHQRTNB",
				VerificationURI:         "https://randomrandom.okta.com/activate",
				VerificationURIComplete: "https://randomrandom.okta.com/activate?user_code=FJHQRTNB",
				ExpiresIn:               3,
				Interval:                1,
			},
			expect:    accessToken{IDToken: "randomness", AccessToken: "b33fid-access-token"},
			id:        "testid",
			status:    http.StatusOK,
			intervals: 3,
		},
		{
			name: "force bad json decode",
			device: Device{
				DeviceCode:              "b33fid",
				UserCode:                "FJHQRTNB",
				VerificationURI:         "https://randomrandom.okta.com/activate",
				VerificationURIComplete: "https://randomrandom.okta.com/activate?user_code=FJHQRTNB",
				ExpiresIn:               3,
				Interval:                1,
			},
			expect:    accessToken{IDToken: "randomness", AccessToken: "b33fid-access-token"},
			id:        "testid",
			status:    http.StatusNotAcceptable,
			intervals: 1,
			err:       ErrAccessTokenJSONDecode,
			errBody:   "{{{ //",
		},
		{
			name: "force bad json decode but 200 status",
			device: Device{
				DeviceCode:              "b33fid",
				UserCode:                "FJHQRTNB",
				VerificationURI:         "https://randomrandom.okta.com/activate",
				VerificationURIComplete: "https://randomrandom.okta.com/activate?user_code=FJHQRTNB",
				ExpiresIn:               3,
				Interval:                1,
			},
			expect:    accessToken{IDToken: "randomness", AccessToken: "b33fid-access-token"},
			id:        "testid",
			status:    http.StatusOK,
			intervals: 1,
			err:       ErrAccessTokenJSONDecode,
			errBody:   "{{{ //",
		},
		{
			name: "expired: good request but times out on device expiration",
			device: Device{
				DeviceCode:              "b33fid",
				UserCode:                "FJHQRTNB",
				VerificationURI:         "https://randomrandom.okta.com/activate",
				VerificationURIComplete: "https://randomrandom.okta.com/activate?user_code=FJHQRTNB",
				ExpiresIn:               2,
				Interval:                1,
			},
			expect:    accessToken{IDToken: "randomness", AccessToken: "b33fid-access-token"},
			id:        "testid",
			status:    http.StatusOK,
			intervals: 3,
			err:       ErrDeviceAuthorizationExpired,
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

			input := &testClientInternalAccessTokenPoll{
				res:        buf.Bytes(),
				clientID:   tt.id,
				deviceCode: tt.device.DeviceCode,
				intervals:  tt.intervals,
				status:     tt.status,
			}
			srv := input.newServerClientInternalAccessTokenPoll()
			defer srv.Close()
			c := Client{uri: srv.URL, id: tt.id}

			// actual test
			token, err := c.accessTokenPoll(tt.device)
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, received: %v", tt.err, err)
			}

			if err != nil {
				return // we got here because we expected an error and it matched
			}

			if token.IDToken != tt.expect.IDToken {
				t.Errorf("expected %v, received %v", tt.expect, token)
			}
		})
	}
}

type testClientInternalSSOAccessToken struct {
	clientID    string
	accessToken string
	idToken     string
	status      int
	res         []byte
}

func newServerClientInternalSSOAccessToken(input testClientInternalSSOAccessToken) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		expectedURI := "/oauth2/v1/token"
		if r.RequestURI != expectedURI {
			w.WriteHeader(http.StatusBadRequest)
			errString := fmt.Sprintf(`{"error":"badURL","error_description":"requested %s and should be %s"}`, r.RequestURI, expectedURI)
			w.Write([]byte(errString))
			return
		}

		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		expectedPayload := url.Values{
			"client_id":            []string{input.clientID},
			"actor_token":          []string{input.accessToken},
			"actor_token_type":     []string{"urn:ietf:params:oauth:token-type:access_token"},
			"subject_token":        []string{input.idToken},
			"subject_token_type":   []string{"urn:ietf:params:oauth:token-type:id_token"},
			"requested_token_type": []string{"urn:okta:oauth:token-type:web_sso_token"},
			"audience":             []string{"urn:okta:apps:" + input.clientID},
			"grant_type":           []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
		}.Encode()

		if string(body) != expectedPayload {
			w.WriteHeader(http.StatusNotAcceptable)
			errString := fmt.Sprintf(`{"error":"badPayload","error_description":"received %s and should be %s"}`, string(body), expectedPayload)
			w.Write([]byte(errString))
			return
		}

		w.WriteHeader(input.status)
		w.Write(input.res)
	}))
}

func TestClientInternalSSOAccessToken(t *testing.T) {
	cases := []struct {
		name    string
		uri     string
		id      string
		token   accessToken
		expect  accessToken
		err     error
		errBody string
		status  int
	}{
		{
			name: "200: good request",
			token: accessToken{
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				AccessToken:  "testaccesstoken",
				Scope:        "openid okta.apps.sso",
				RefreshToken: "",
				IDToken:      "testidtoken",
			},
			expect: accessToken{
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				AccessToken:  "testaccesstoken-answer",
				Scope:        "openid okta.apps.sso",
				RefreshToken: "",
				IDToken:      "testidtoken-answer",
			},
			id:     "testid",
			status: http.StatusOK,
		},
		{
			name: "forcing json decode error",
			token: accessToken{
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				AccessToken:  "testaccesstoken",
				Scope:        "openid okta.apps.sso",
				RefreshToken: "",
				IDToken:      "testidtoken",
			},
			errBody: "{{{{ //",
			id:      "testid",
			status:  http.StatusNotAcceptable,
			err:     ErrSSOJSONDecode,
		},
		{
			name: "forcing json decode error but 200 status",
			token: accessToken{
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				AccessToken:  "testaccesstoken",
				Scope:        "openid okta.apps.sso",
				RefreshToken: "",
				IDToken:      "testidtoken",
			},
			expect:  accessToken{},
			errBody: "{{{ //",
			id:      "testid",
			status:  http.StatusOK,
			err:     ErrSSOJSONDecode,
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

			input := testClientInternalSSOAccessToken{
				res:         buf.Bytes(),
				clientID:    tt.id,
				status:      tt.status,
				accessToken: tt.token.AccessToken,
				idToken:     tt.token.IDToken,
			}
			srv := newServerClientInternalSSOAccessToken(input)
			defer srv.Close()
			c := Client{uri: srv.URL, id: tt.id, appID: tt.id}

			// actual test
			sso, err := c.ssoAccessToken(tt.token)
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, received: %v", tt.err, err)
			}

			if err != nil {
				return // we got here because we expected an error and it matched
			}

			if sso.IDToken != tt.expect.IDToken {
				t.Errorf("expected %v, received %v", tt.expect, sso)
			}
		})
	}
}
