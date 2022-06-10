package okta

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/html"
)

// ErrSAMLRequest is returned when there's a failure related to the request.
// The response is not the expected response or it failed to complete the
// request.
var ErrSAMLRequest = errors.New("saml request")

// ErrSAMLResponse is returned when there's a failure reading the response from
// the request.
var ErrSAMLResponse = errors.New("saml response")

// getSAML returns the SAML credential values after extracting it from the SSO
// login endpoint of Okta. Returns non-nil error if the request goes wrong,
// there's an issue reading the response body or can't find a SAML response to
// extract.
func (c Client) getSAML(sso accessToken) (string, error) {
	uri := fmt.Sprintf("%s/login/token/sso?token=%s", c.uri, sso.AccessToken)
	resp, err := http.Get(uri)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSAMLRequest, err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("saml: reading error response: %w", err)
		}
		defer resp.Body.Close()

		return "", fmt.Errorf("%w: status %d: %s", ErrSAMLRequest, resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSAMLResponse, err)
	}

	saml, err := extractSAML(body)
	if err != nil {
		return "", err
	}

	return saml, nil
}

// ErrNoSAMLResponseFound is returned when the HTML from the SSO login has no
// SAMLResponse tag in its body.
var ErrNoSAMLResponseFound = errors.New("no SAMLResponse found")

// ErrParsingHTML is returned when the HTML from the SSO login can't be parsed.
// This probably will never happen because the HTML parser returns early on any
// error, without error. And the only error that handles is io.EOF, as an exit
// as well.
var ErrParsingHTML = errors.New("parsing html")

// extractSAML returns a base64 SAML Response from a hidden HTML input tag.
// Returns non-nil error in case the HTML can't be parsed or can't find a
// SAMLResponse HTML ElementNode with ErrNoSAMLResponseFound.
func extractSAML(src []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(src))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParsingHTML, err)
	}

	response := ""

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			present := false

			// finding the right html.ElementNode.
			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "SAMLResponse" {
					present = true
					break
				}
			}

			if present {
				// finding the attribute value and its value if in the right
				// html.ElementNode.
				for _, attr := range n.Attr {
					if attr.Key == "value" {
						response = attr.Val
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if response == "" {
		return "", ErrNoSAMLResponseFound
	}

	return response, nil
}
