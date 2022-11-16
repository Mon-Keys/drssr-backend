package client

import (
	"io"
	"net/http"
	"time"
)

type Client struct {
	url            string
	secret         string
	client         *http.Client
	contextTimeout time.Duration
}

func New(url string, secret string, timeout time.Duration) *Client {
	client := &Client{
		url:    url,
		secret: secret,
		client: &http.Client{
			Timeout: timeout,
		},
	}

	return client
}

func (c *Client) readBody(readCloser io.ReadCloser) ([]byte, error) {
	defer readCloser.Close()

	body, err := io.ReadAll(readCloser)
	if err != nil {
		return nil, err
	}

	return body, nil
}
