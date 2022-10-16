package gomodanalysis

import (
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// GithubClient defines the HTTP client interface.
type GithubClient interface {
	// Fetch sends a GET request and return raw data.
	Fetch(url string) (*Response, error)

	// GraphQL sends a graphQL request.
	GraphQL(query string) (*Response, error)
}

// Client defines the HTTP client to fetch html pages from GitHub.
type Client struct {
	cfg Configuration
}

type Response struct {
	Body       io.ReadCloser
	StatusCode int
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Configuration Client configurations.
type Configuration struct {
	HTTPClient httpClient
	Cookies    []*http.Cookie
	Backoff    *Backoff
	APIToken   string
	GraphQLURL string
}

// NewClient initialises a new HTTP Client.
func NewClient(cfg Configuration) *Client {
	var jar http.CookieJar
	if cfg.Cookies != nil {
		jar, _ = cookiejar.New(nil)
		u, _ := url.Parse("https://github.com")
		jar.SetCookies(u, cfg.Cookies)
	}

	if cfg.GraphQLURL == "" {
		cfg.GraphQLURL = "https://api.github.com/graphql"
	}

	if cfg.Backoff == nil {
		cfg.Backoff = &Backoff{
			MaxDelay: 10,
			MaxSteps: 10,
		}
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Jar:     jar,
			Timeout: 5 * time.Second,
		}
	}

	return &Client{
		cfg: cfg,
	}
}

func (c *Client) do(req *http.Request) (*Response, error) {
	if c.cfg.APIToken != "" {
		req.Header.Set("Authorization", "bearer "+c.cfg.APIToken)
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	r := Response{Body: resp.Body, StatusCode: resp.StatusCode}

	if r.StatusCode > 209 {
		d, err := c.cfg.Backoff.LinearDelay()
		if err != nil {
			return &r, err
		}
		time.Sleep(d)
		return c.do(req)
	}

	return &r, nil
}

func (c *Client) Fetch(url string) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) GraphQL(query string) (*Response, error) {
	req, err := http.NewRequest("POST", c.cfg.GraphQLURL, strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

type Backoff struct {
	MaxDelay time.Duration
	MaxSteps int64

	step int64
	mu   sync.Mutex
}

func (b *Backoff) LinearDelay() (time.Duration, error) {
	b.mu.Lock()
	b.step++
	b.mu.Unlock()

	d := b.MaxDelay.Microseconds() / b.MaxSteps * b.step
	if d > b.MaxDelay.Microseconds() {
		return 0, errors.New("max retry delay has been reached")
	}
	return time.Duration(d), nil
}
