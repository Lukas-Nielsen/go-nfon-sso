package nfon

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func generateUnique(length int) (string, error) {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateCodeVerifier(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	base64url := base64.RawURLEncoding.EncodeToString(bytes)

	return base64url, nil
}

func generateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return codeChallenge
}

func getCodeFromURL(rawURL string) string {
	var fragment string

	if idx := strings.Index(rawURL, "#"); idx != -1 {
		fragment = rawURL[idx+1:]
	} else if idx := strings.Index(rawURL, "?"); idx != -1 {
		fragment = rawURL[idx+1:]
	} else {
		return ""
	}

	values, err := url.ParseQuery(fragment)
	if err != nil {
		return ""
	}

	return values.Get("code")
}

func getFormActionFromBody(htmlBody string, id string) (string, error) {
	// Create a reader from the HTML body string
	reader := strings.NewReader(htmlBody)

	// Parse the HTML content
	doc, err := html.Parse(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Traverse the HTML nodes to find the form element with id <id>
	var formAction string
	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		// Look for form element with id="<id>"
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == id {
					// Get the action attribute
					for _, attr := range n.Attr {
						if attr.Key == "action" {
							formAction = attr.Val
							return
						}
					}
				}
			}
		}
		// Traverse child nodes recursively
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	// Start traversing from the root node
	visitNode(doc)

	if formAction == "" {
		return "", fmt.Errorf("form with id 'kc-form-login' not found")
	}

	return formAction, nil
}
