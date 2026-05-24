package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"scamshield/internal/domain"
)

type MLClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPMLClient(baseURL string) *MLClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &MLClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 750 * time.Millisecond},
	}
}

func (c *MLClient) ScoreText(ctx context.Context, req domain.ModelScoreRequest) (domain.ModelScoreResponse, error) {
	return c.score(ctx, "/internal/model/score-text", req)
}

func (c *MLClient) ScoreURL(ctx context.Context, req domain.ModelScoreRequest) (domain.ModelScoreResponse, error) {
	return c.score(ctx, "/internal/model/score-url", req)
}

func (c *MLClient) Metadata(ctx context.Context) (map[string]any, error) {
	if c == nil {
		return nil, errors.New("ml client is disabled")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/model/metadata", nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errors.New("ml metadata request failed")
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *MLClient) score(ctx context.Context, path string, req domain.ModelScoreRequest) (domain.ModelScoreResponse, error) {
	if c == nil {
		return domain.ModelScoreResponse{}, errors.New("ml client is disabled")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return domain.ModelScoreResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return domain.ModelScoreResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return domain.ModelScoreResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return domain.ModelScoreResponse{}, errors.New("ml score request failed")
	}
	var payload domain.ModelScoreResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return domain.ModelScoreResponse{}, err
	}
	return payload, nil
}
