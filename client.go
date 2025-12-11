package nfon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

const (
	authBaseURL         = "https://sso.cloud-cfg.com/realms/login/protocol/openid-connect"
	formLogin           = "kc-form-login"
	formOTP             = "kc-otp-login-form"
	scopeOpenID         = "openid"
	userAgent           = "go-nfon-sso"
	defaultMaxRedirects = 20
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
	portalBaseURL string
	clientId      string
	codeVerifier  string
	token         Token
	client        *resty.Client
	requestCount  int32
}

func NewClient(portalBaseURL, clientId string) (*Client, error) {
	c := &Client{
		portalBaseURL: portalBaseURL,
		clientId:      clientId,
	}
	c.setup()
	return c, nil
}

func (c *Client) SetPortalBaseURL(portalBaseURL string) *Client {
	c.portalBaseURL = portalBaseURL
	return c
}

func (c *Client) SetClientId(clientId string) *Client {
	c.clientId = clientId
	return c
}

func (c *Client) Login(username, password string) (string, error) {
	state, _ := generateUnique(32)
	nonce, _ := generateUnique(32)
	c.codeVerifier, _ = generateCodeVerifier(43)
	codeChallenge := generateCodeChallenge(c.codeVerifier)

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"client_id":             c.clientId,
			"redirect_uri":          c.portalBaseURL,
			"state":                 state,
			"response_mode":         "fragment",
			"response_type":         "code",
			"scope":                 scopeOpenID,
			"nonce":                 nonce,
			"code_challenge":        codeChallenge,
			"code_challenge_method": "S256",
		}).
		Get(authBaseURL + "/auth")
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("login failed: %s", resp.Status())
	}

	formUrl, err := getFormActionFromBody(resp.String(), formLogin)
	if err != nil {
		return "", fmt.Errorf("failed to parse login form: %w", err)
	}

	// username
	c.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, err = c.client.R().
		SetFormData(map[string]string{"username": username}).
		Post(formUrl)
	c.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(defaultMaxRedirects))
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("username step failed: %s", resp.Status())
	}

	formUrl, err = getFormActionFromBody(resp.String(), formLogin)
	if err != nil {
		return "", fmt.Errorf("failed to parse password form: %w", err)
	}

	// username + password
	c.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, err = c.client.R().
		SetFormData(map[string]string{
			"username":     username,
			"password":     password,
			"rememberMe":   "on",
			"credentialId": "",
		}).
		Post(formUrl)
	c.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(defaultMaxRedirects))
	if err != nil {
		return "", err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		formUrl, err = getFormActionFromBody(resp.String(), formOTP)
		if err != nil {
			return "", fmt.Errorf("failed to parse OTP form: %w", err)
		}
		return formUrl, nil
	case http.StatusFound:
		return "", c.fetchToken(getCodeFromURL(resp.Header().Get("Location")))
	default:
		return "", fmt.Errorf("unexpected response: %s", resp.Status())
	}
}

func (c *Client) OTP(url, otp string) (string, error) {
	c.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"otp":   otp,
			"login": "Loggen+Sie+sich+ein",
		}).
		Post(url)
	c.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(defaultMaxRedirects))
	if err != nil && !strings.Contains(err.Error(), "auto redirect is disabled") {
		return "", err
	}
	if resp.StatusCode() == http.StatusFound {
		return "", c.fetchToken(getCodeFromURL(resp.Header().Get("Location")))
	}

	formUrl, err := getFormActionFromBody(resp.String(), formOTP)
	if err != nil {
		return "", fmt.Errorf("OTP failed and failed to parse OTP form: %w", err)
	}
	return formUrl, fmt.Errorf("OTP failed: %s", resp.Status())
}

func (c *Client) Logout() {
	if c != nil {
		c.client.R().
			SetQueryParams(map[string]string{
				"client_id":                c.clientId,
				"post_logout_redirect_uri": c.portalBaseURL,
				"id_token_hint":            c.token.IDToken,
			}).
			Get(authBaseURL + "/logout")

		// Cookies löschen
		c.client.SetCookieJar(nil)
	}
}

func (c *Client) setup() {
	c.client = resty.New().
		SetHeader("User-Agent", userAgent)
}

func (c *Client) fetchToken(code string) error {
	resp, err := c.client.R().
		SetResult(&c.token).
		SetFormData(map[string]string{
			"code":          code,
			"grant_type":    "authorization_code",
			"client_id":     c.clientId,
			"redirect_uri":  c.portalBaseURL,
			"code_verifier": c.codeVerifier,
		}).
		Post(authBaseURL + "/token")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("token fetch failed: %s", resp.Status())
	}
	return nil
}

func (c *Client) RefreshToken() error {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": c.token.RefreshToken,
			"client_id":     c.clientId,
		}).
		SetResult(&c.token).
		Post(authBaseURL + "/token")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("token refresh failed: %s", resp.Status())
	}
	return nil
}

func (c *Client) GetToken() Token      { return c.token }
func (c *Client) SetToken(token Token) { c.token = token }
func (c *Client) GetRequestCount() int { return int(atomic.LoadInt32(&c.requestCount)) }
func (c *Client) ResetRequestCount()   { atomic.StoreInt32(&c.requestCount, 0) }

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

// ---- Generische Request-Methode ----
func (c *Client) doRequest(method, uri string, payload any, query, header map[string]string, result any) (*resty.Response, error) {
	atomic.AddInt32(&c.requestCount, 1)

	req := c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header)

	if payload != nil {
		req.SetBody(payload)
	}
	if result != nil {
		req.SetResult(result)
	}

	switch method {
	case http.MethodGet:
		return req.Get(uri)
	case http.MethodPost:
		return req.Post(uri)
	case http.MethodPut:
		return req.Put(uri)
	case http.MethodPatch:
		return req.Patch(uri)
	case http.MethodDelete:
		return req.Delete(uri)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func (c *Client) UploadFile(uri string, query, header map[string]string, fieldName string, fileName string, mimeType string, formData map[string]string) (*resty.Response, error) {
	atomic.AddInt32(&c.requestCount, 1)

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return c.client.R().
		SetAuthScheme(c.token.TokenType).
		SetAuthToken(c.token.AccessToken).
		SetQueryParams(query).
		SetHeaders(header).
		SetMultipartField(
			fieldName,
			filepath.Base(fileName),
			mimeType,
			file,
		).
		SetFormData(formData).
		Post(uri)
}

// Wrapper
func (c *Client) Get(uri string, query, header map[string]string) (*resty.Response, error) {
	return c.doRequest(http.MethodGet, uri, nil, query, header, nil)
}
func (c *Client) Post(uri string, payload any, query, header map[string]string) (*resty.Response, error) {
	return c.doRequest(http.MethodPost, uri, payload, query, header, nil)
}
func (c *Client) Put(uri string, payload any, query, header map[string]string) (*resty.Response, error) {
	return c.doRequest(http.MethodPut, uri, payload, query, header, nil)
}
func (c *Client) Patch(uri string, payload any, query, header map[string]string) (*resty.Response, error) {
	return c.doRequest(http.MethodPatch, uri, payload, query, header, nil)
}
func (c *Client) Delete(uri string, query, header map[string]string) (*resty.Response, error) {
	return c.doRequest(http.MethodDelete, uri, nil, query, header, nil)
}

func (c *Client) GetPortalApi(uri string, query, header map[string]string) (Response, error) {
	var result Response
	_, err := c.doRequest(http.MethodGet, uri, nil, query, header, &result)
	return result, err
}
func (c *Client) PostPortalApi(uri string, payload any, query, header map[string]string) (Response, error) {
	var result Response
	_, err := c.doRequest(http.MethodPost, uri, payload, query, header, &result)
	return result, err
}
