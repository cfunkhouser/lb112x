package lb112x

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type GeneralInfo struct {
	Model             string `json:"model"`
	VersionMajor      int    `json:"verMajor"`
	VersionMinor      int    `json:"verMinor"`
	HardwareVersion   string `json:"HWversion"`
	FirmwareVersion   string `json:"FWversion"`
	AppVersion        string `json:"appVersion"`
	WebAppVersion     string `json:"webAppVersion"`
	APIVersion        string `json:"apiVersion"`
	BootloaderVersion string `json:"BLversion"`
	IMEI              string `json:"IMEI"`
	Temperature       int    `json:"devTemperature"`
}

type PowerInfo struct {
	DeviceTempCritical bool   `json:"deviceTempCritical"`
	ResetRequired      string `json:"resetRequired"`
}

type SessionInfo struct {
	UserRole      string `json:"userRole"`
	Language      string `json:"lang"`
	SecurityToken string `json:"secToken"`
}

type SignalStrenthInfo struct {
	RSSI int `json:"rssi"`
	RSCP int `json:"rscp"`
	ECIO int `json:"ecio"`
	RSRP int `json:"rsrp"`
	RSRQ int `json:"rsrq"`
	Bars int `json:"bars"`
	SINR int `json:"SINR"`
}

type WWANInfo struct {
	IP                     string            `json:"IP"`
	IPv6                   string            `json:"IPv6"`
	RegisterNetworkDisplay string            `json:"registerNetworkDisplay"`
	SignalStrenth          SignalStrenthInfo `json:"signalStrength"`
}

type APIModel struct {
	General GeneralInfo `json:"general"`
	Power   PowerInfo   `json:"power"`
	Session SessionInfo `json:"session"`
	WWAN    WWANInfo    `json:"wwan"`
}

// Client for communicating with the LB112X API.
type Client struct {
	c             *http.Client
	url, password string
}

// Poll the LB112X API.
func (c *Client) Poll() (*APIModel, error) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/model.json?internalapi=1", c.url), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var model APIModel
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, err
	}
	return &model, nil
}

// GetSession information from the LB112X API.
func (c *Client) GetSession() (*SessionInfo, error) {
	model, err := c.Poll()
	if err != nil {
		return nil, err
	}
	return &(model.Session), nil
}

var ErrFailedAuthentication = errors.New("failed to log in")

// Authenticate with the API.
func (c *Client) Authenticate() error {
	sess, err := c.GetSession()
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Add("token", sess.SecurityToken)
	form.Add("session.password", c.password)
	r, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/Forms/config", c.url), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.c.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if int(resp.StatusCode/100) != 2 {
		return fmt.Errorf("%w: %v %v", ErrFailedAuthentication, resp.StatusCode, resp.Status)
	}
	return nil
}

// DefaultTransport for HTTP requests to the LB112X API. Useful for wrapping the
// transport from outside the library.
func DefaultTransport() http.RoundTripper {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Dial: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}
}

type initOptions struct {
	ClientTimeout time.Duration
	CookieJar     http.CookieJar
	Transport     http.RoundTripper
}

func defaultInitOptions() *initOptions {
	jar, _ := cookiejar.New(nil)
	return &initOptions{
		CookieJar:     jar,
		ClientTimeout: 2 * time.Second,
		Transport:     DefaultTransport(),
	}
}

// Option for the creation of the LB112X API Client.
type Option func(*initOptions)

// WithRoundTripper for HTTP requests to the LB112X API.
func WithRoundTripper(rt http.RoundTripper) Option {
	return func(o *initOptions) {
		o.Transport = rt
	}
}

// WithClientTimeout while connecting to the LB112X API.
func WithClientTimeout(ttl time.Duration) Option {
	return func(o *initOptions) {
		o.ClientTimeout = ttl
	}
}

// WithCookieJar for use by the HTTP client communicating with the LB112X API.
func WithCookieJar(jar http.CookieJar) Option {
	return func(o *initOptions) {
		o.CookieJar = jar
	}
}

// New LB112X API Client.
func New(url, password string, opts ...Option) *Client {
	iopts := defaultInitOptions()
	for _, opt := range opts {
		opt(iopts)
	}
	return &Client{
		c: &http.Client{
			Jar:       iopts.CookieJar,
			Transport: iopts.Transport,
			Timeout:   iopts.ClientTimeout,
		},
		url:      url,
		password: password,
	}
}
