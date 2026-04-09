package facepp

import (
	"bytes"
	"context"
	"encoding/json"

	"mime/multipart"
	"net/http"
)

type Client interface {
	CompareFaces(ctx context.Context, selfie []byte, document []byte) (*CompareResponse, error)
}

type client struct {
	apiKey    string
	apiSecret string
	http      *http.Client
}

func NewClient(apiKey, apiSecret string) Client {
	return &client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		http:      &http.Client{},
	}
}

type CompareResponse struct {
	Confidence float64 `json:"confidence"`
	Thresholds struct {
		Threshold80 float64 `json:"1e-3"`
	} `json:"thresholds"`
}

func (c *client) CompareFaces(ctx context.Context, selfie []byte, document []byte) (*CompareResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("api_key", c.apiKey)
	_ = writer.WriteField("api_secret", c.apiSecret)

	part1, _ := writer.CreateFormFile("image_file1", "selfie.jpg")
	part1.Write(selfie)

	part2, _ := writer.CreateFormFile("image_file2", "document.jpg")
	part2.Write(document)

	writer.Close()

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api-us.faceplusplus.com/facepp/v3/compare",
		body,
	)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
