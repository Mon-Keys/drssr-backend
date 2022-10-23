package cutter

import (
	"bytes"
	"context"
	"drssr/internal/pkg/ctx_utils"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

type UploadImgArgs struct {
	FileHeader multipart.FileHeader
	File       []byte
}

type UploadImgRes struct {
	Img      string `json:"img"`
	ImgPath  string `json:"img_path"`
	Mask     string `json:"mask"`
	MaskPath string `json:"mask_path"`
}

func (c *Client) UploadImg(ctx context.Context, args *UploadImgArgs) (*UploadImgRes, error) {
	// ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	// defer cancel()

	reqBody := &bytes.Buffer{}
	w := multipart.NewWriter(reqBody)
	part, err := w.CreatePart(args.FileHeader.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to create part: %w", err)
	}
	part.Write(args.File)
	w.Close()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url+"/upload",
		reqBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
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

	var data *UploadImgRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return data, nil
}
