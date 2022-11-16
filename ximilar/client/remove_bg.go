package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type RemoveBGReq struct {
	Records []RecordsReqStruct `json:"records"`
}

type RecordsReqStruct struct {
	URL             string `json:"_url"`
	Base64          string `json:"_base64"`
	BinaryMask      bool   `json:"binary_mask"`
	WhiteBackground bool   `json:"white_background"`
}

type RemoveBGRes struct {
	Status  StatusStruct       `json:"status"`
	Records []RecordsResStruct `json:"records"`
}

type StatusStruct struct {
	Code  int    `json:"code"`
	Text  string `json:"text"`
	ReqID string `json:"request_id"`
}

type RecordsResStruct struct {
	Status    StatusStruct `json:"_status"`
	ID        string       `json:"_id"`
	Width     int          `json:"_width"`
	Height    int          `json:"_height"`
	OutputURL string       `json:"_output_url"`
}

func (c *Client) RemoveBGPrecise(ctx context.Context, args *RemoveBGReq) (*RemoveBGRes, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal failed with: %s", err)
	}

	// ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	// defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url+"/removebg/precise/removebg",
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// if reqID := ctx_utils.GetReqID(ctx); reqID != "" {
	// 	req.Header.Set("X-Request-ID", reqID)
	// }
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.secret))

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

	var data *RemoveBGRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return data, nil
}

func (c *Client) RemoveBGFast(ctx context.Context, args *RemoveBGReq) (*RemoveBGRes, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal failed with: %s", err)
	}

	// ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	// defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url+"/removebg/fast/removebg",
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// if reqID := ctx_utils.GetReqID(ctx); reqID != "" {
	// 	req.Header.Set("X-Request-ID", reqID)
	// }
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.secret))

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

	var data *RemoveBGRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return data, nil
}
