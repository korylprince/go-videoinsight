package videoinsight

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// AuthenticationError is an API authentication error
type AuthenticationError string

func (e AuthenticationError) Error() string {
	return string(e)
}

// UnknownError is an unknown HTTP error
type UnknownError string

func (e UnknownError) Error() string {
	return fmt.Sprintf("unknown error: %s", string(e))
}

// Client is a VideoInsight API client
type Client struct {
	proto string
	host  string
	port  int
	token string
}

// NewClient returns an unauthorized client
func NewClient(proto, host string, port int) *Client {
	return &Client{proto: proto, host: host, port: port}
}

// Authenticate attempts to authenticate the client and stores the authentication token for future requests
// If username or password is invalid, AuthenticationError is returned
func (c *Client) Authenticate(username, password string, expireSeconds int) error {
	authURL := &url.URL{
		Scheme:   c.proto,
		Host:     fmt.Sprintf("%s:%d", c.host, c.port),
		Path:     "/api/v1/authenticate",
		RawQuery: url.Values{"name": []string{username}, "expiresInSeconds": []string{strconv.Itoa(expireSeconds)}}.Encode(),
	}
	resp, err := http.PostForm(authURL.String(), url.Values{"password": []string{password}})
	if err != nil {
		return fmt.Errorf("could not complete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AuthenticationError(resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read body: %w", err)
	}

	token, err := url.QueryUnescape(string(buf))
	if err != nil {
		return fmt.Errorf("could not parse token: %w", err)
	}
	c.token = strings.Trim(token, `"`)

	return nil
}
