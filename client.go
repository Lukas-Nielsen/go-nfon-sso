package nfon

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type Client struct {
	portalBaseUrl string
	clientId      string
	codeVerifier  string
	token         Token
	client        *resty.Client
}

func NewClient(portalBaseUrl string, clientId string) (*Client, error) {
	c := Client{
		portalBaseUrl: portalBaseUrl,
		clientId:      clientId,
	}

	c = *c.setup()

	return &c, nil
}

func (c *Client) SetPortalBaseUrl(portalBaseUrl string) *Client {
	c.portalBaseUrl = portalBaseUrl
	return c
}

func (c *Client) SetClientId(clientId string) *Client {
	c.clientId = clientId
	return c
}

func (c *Client) Login(username string, password string) (string, error) {
	state, _ := generateUnique(32)
	nonce, _ := generateUnique(32)
	c.codeVerifier, _ = generateCodeVerifier(43)
	codeChallenge := generateCodeChallenge(c.codeVerifier)

	// get login form
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"client_id":             c.clientId,
			"redirect_uri":          c.portalBaseUrl,
			"state":                 state,
			"response_mode":         "fragment",
			"response_type":         "code",
			"scope":                 "openid",
			"nonce":                 nonce,
			"code_challenge":        codeChallenge,
			"code_challenge_method": "S256",
		}).
		Get("https://sso.cloud-cfg.com/realms/login/protocol/openid-connect/auth")

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("%s", resp.String())
	}

	formUrl, _ := getFormActionFromBody(resp.String(), "kc-form-login")

	// login
	c.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, err = c.client.R().
		SetFormData(map[string]string{
			"username":     username,
			"password":     password,
			"rememberMe":   "on",
			"credentialId": "",
		}).
		Post(formUrl)
	c.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))

	if err != nil {
		return "", err
	}

	if resp.StatusCode() == 200 {
		// get otp form url
		formUrl, _ = getFormActionFromBody(resp.String(), "kc-otp-login-form")
		return formUrl, nil
	} else if resp.StatusCode() == 302 {
		// fetch access token
		return "", c.fetchToken(getCodeFromURL(resp.Header().Get("Location")))
	}

	return "", fmt.Errorf("%s", resp.String())
}

func (c *Client) OTP(url string, otp string) error {
	// do otp
	c.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"otp":   otp,
			"login": "Loggen+Sie+sich+ein",
		}).
		Post(url)
	c.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))

	if err != nil && !strings.Contains(err.Error(), "auto redirect is disabled") {
		return err
	}

	if resp.StatusCode() == 302 {
		// fetch access token
		return c.fetchToken(getCodeFromURL(resp.Header().Get("Location")))
	}

	return fmt.Errorf("%s", resp.String())
}

func (c *Client) setup() *Client {
	c.client = resty.New().
		SetHeader("User-Agent", "go-nfon-sso")

	return c
}

func (c *Client) fetchToken(code string) error {
	resp, err := c.client.R().
		SetResult(&c.token).
		SetFormData(map[string]string{
			"code":          code,
			"grant_type":    "authorization_code",
			"client_id":     c.clientId,
			"redirect_uri":  c.portalBaseUrl,
			"code_verifier": c.codeVerifier,
		}).
		Post("https://sso.cloud-cfg.com/realms/login/protocol/openid-connect/token")
	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		return nil
	}
	return fmt.Errorf("%s", resp.String())
}

func (c *Client) RefreshToken() error {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": c.token.RefreshToken,
			"client_id":     c.clientId,
		}).
		SetResult(&c.token).
		Post("https://sso.cloud-cfg.com/realms/login/protocol/openid-connect/token")

	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		return nil
	}
	return fmt.Errorf("%s", resp.String())
}

func (c *Client) GetToken() Token {
	return c.token
}

func (c *Client) SetToken(token Token) {
	c.token = token
}

func (c *Client) TokenFromJsonFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &c.token)
}

func (c *Client) TokenToJsonFile(path string) error {

	data, err := json.Marshal(c.token)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0666)
}

func (c *Client) Get(uri string, query map[string]string, header map[string]string) (*resty.Response, error) {
	return c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header).
		Get(uri)
}

func (c *Client) Delete(uri string, query map[string]string, header map[string]string) (*resty.Response, error) {
	return c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header).
		Delete(uri)
}

func (c *Client) Post(uri string, payload any, query map[string]string, header map[string]string) (*resty.Response, error) {
	return c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header).
		SetBody(payload).
		Post(uri)
}

func (c *Client) Put(uri string, payload any, query map[string]string, header map[string]string) (*resty.Response, error) {
	return c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header).
		SetBody(payload).
		Put(uri)
}

func (c *Client) Patch(uri string, payload any, query map[string]string, header map[string]string) (*resty.Response, error) {
	return c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header).
		SetBody(payload).
		Patch(uri)
}
