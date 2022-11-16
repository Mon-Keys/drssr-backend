package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DominantColotReq struct {
	ColorNames bool                           `json:"color_names"`
	Records    []DominantColotRecordReqStruct `json:"records"`
}

type DominantColotRecordReqStruct struct {
	URL    string `json:"_url"`
	Base64 string `json:"_base64"`
}

type DominantColorRes struct {
	Status  StatusStruct                    `json:"status"`
	Records []DominantColorRecordsResStruct `json:"records"`
}

type DominantColorRecordsResStruct struct {
	Status         StatusStruct         `json:"_status"`
	ID             string               `json:"_id"`
	Width          int                  `json:"_width"`
	Height         int                  `json:"_height"`
	DominantColors DominantColorsStruct `json:"_dominant_colors"`
}

type DominantColorsStruct struct {
	Percentages       []float32 `json:"percentages"`
	ColorNames        []string  `json:"color_names"`
	ColorNamesPantone []string  `json:"color_names_pantone"`
}

func (c *Client) DominantColorProduct(ctx context.Context, args *DominantColotReq) (*DominantColorRes, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal failed with: %s", err)
	}

	// ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	// defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url+"/dom_colors/product/v2/dominantcolor",
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

	var data *DominantColorRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return data, nil
}
