package okta

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/html"
)

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
