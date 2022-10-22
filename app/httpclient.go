package gomodanalysis

import (
	"errors"
	"io"
	"log"
	"net"
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
	Verbose    bool
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
			MaxDelay: 1 * time.Minute,
			MaxSteps: 60,
		}
	}

	if cfg.HTTPClient == nil {
		t := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		cfg.HTTPClient = &http.Client{
			Jar:       jar,
			Timeout:   10 * time.Second,
			Transport: t,
		}
	}

	return &Client{
		cfg: cfg,
	}
}

func (c *Client) doAfterLinerDelay(req *http.Request) (*http.Response, error) {
	d, err := c.cfg.Backoff.LinearDelay()
	if err != nil {
		return nil, err
	}
	if c.cfg.Verbose && d > 0 {
		log.Printf("delay for %v sec. and retry", d.Seconds())
	}
	time.Sleep(d)
	return c.cfg.HTTPClient.Do(req)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.cfg.APIToken != "" {
		req.Header.Set("Authorization", "bearer "+c.cfg.APIToken)
	}

	for {
		resp, err := c.doAfterLinerDelay(req)
		switch err.(type) {
		case nil:
			if resp.StatusCode > 209 {
				c.cfg.Backoff.UpCounter()
				continue
			}
			c.cfg.Backoff.Reset()
			return resp, nil
		case (*url.Error):
			if e, ok := err.(*url.Error); ok {
				if e.Err == io.EOF || e.Err == io.ErrUnexpectedEOF {
					c.cfg.Backoff.UpCounter()
					continue
				}
			}
		case TimeoutError:
			c.cfg.Backoff.Reset()
			return nil, err
		default:
			c.cfg.Backoff.Reset()
			return nil, errors.New("c.do: error: " + err.Error())
		}
	}
}

func (c *Client) Fetch(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New(`c.Fetch: http.NewRequest("GET", url, nil) error: ` + err.Error())
	}
	return c.do(req)
}

func (c *Client) GraphQL(query string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.cfg.GraphQLURL, strings.NewReader(query))
	if err != nil {
		return nil, errors.New(
			`c.GraphQL: http.NewRequest("POST", c.cfg.GraphQLURL, strings.NewReader(query)) error: ` + err.Error(),
		)
	}
	return c.do(req)
}

type Backoff struct {
	MaxDelay time.Duration
	MaxSteps int64

	step int64
	mu   sync.Mutex
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "max retry delay has been reached"
}

func (b *Backoff) LinearDelay() (time.Duration, error) {
	if b.step > b.MaxSteps {
		return 0, TimeoutError{}
	}
	return time.Duration(b.MaxDelay.Nanoseconds() / b.MaxSteps * b.step), nil
}

func (b *Backoff) UpCounter() {
	b.mu.Lock()
	b.step++
	b.mu.Unlock()
}

func (b *Backoff) Reset() {
	b.mu.Lock()
	b.step = 0
	b.mu.Unlock()
}
