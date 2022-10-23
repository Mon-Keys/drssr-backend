package similarity

import (
	"bytes"
	"context"
	"drssr/internal/pkg/ctx_utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	url            string
	client         *http.Client
	contextTimeout time.Duration
}

func New(url string, timeout time.Duration) *Client {
	client := &Client{
		url: url,
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

// {
// 	"image" : "hui.png",
// 	"images": {
// 	 "1": "ahah.png",
// 	 "2": "mishaloh.png",
// 	}
// }

type CheckSimilarityArgs struct {
	CheckedImage   string            `json:"image"`
	CheckingImages map[uint64]string `json:"images"`
}

type CheckSimilarityRes struct {
	Similarity map[uint64]int `json:"similarity"`
}

func (c *Client) CheckSimilarity(ctx context.Context, args *CheckSimilarityArgs) (*CheckSimilarityRes, error) {
	// ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	// defer cancel()

	reqBody, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url+"/similarity",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if reqID := ctx_utils.GetReqID(ctx); reqID != "" {
		req.Header.Set("X-Request-ID", reqID)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()

	body, err := c.readBody(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with: status=%s, url='%s', body='%s'", res.Status, req.URL.String(), body)
	}

	var data *CheckSimilarityRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return data, nil
}
