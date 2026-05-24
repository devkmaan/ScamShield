package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"scamshield/internal/domain"
)

type GenAIClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPGenAIClient(baseURL string) *GenAIClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &GenAIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: genAIHTTPTimeout()},
	}
}

func genAIHTTPTimeout() time.Duration {
	value := strings.TrimSpace(os.Getenv("GENAI_CLIENT_TIMEOUT_MS"))
	if value == "" {
		return 35 * time.Second
	}
	ms, err := strconv.Atoi(value)
	if err != nil || ms <= 0 {
		return 35 * time.Second
	}
	return time.Duration(ms) * time.Millisecond
}

func (c *GenAIClient) NormalizeInput(ctx context.Context, req domain.GenAINormalizeRequest) (domain.GenAINormalizeResponse, error) {
	var payload domain.GenAINormalizeResponse
	err := c.post(ctx, "/internal/genai/normalize-input", req, &payload)
	return payload, err
}

func (c *GenAIClient) Render(ctx context.Context, req domain.GenAIRenderRequest) (domain.GenAIRenderResponse, error) {
	var payload domain.GenAIRenderResponse
	err := c.post(ctx, "/internal/genai/render", req, &payload)
	return payload, err
}

func (c *GenAIClient) Chat(ctx context.Context, req domain.GenAIChatRequest) (domain.GenAIChatResponse, error) {
	var payload domain.GenAIChatResponse
	err := c.post(ctx, "/internal/genai/chat", req, &payload)
	return payload, err
}

func (c *GenAIClient) UIBundle(ctx context.Context, req domain.UIBundleRequest) (domain.UIBundleResponse, error) {
	var payload domain.UIBundleResponse
	err := c.post(ctx, "/internal/genai/ui-bundle", req, &payload)
	return payload, err
}

func (c *GenAIClient) Metadata(ctx context.Context) (map[string]any, error) {
	if c == nil {
		return nil, errors.New("genai client is disabled")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/genai/metadata", nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errors.New("genai metadata request failed")
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *GenAIClient) post(ctx context.Context, path string, req any, out any) error {
	if c == nil {
		return errors.New("genai client is disabled")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errors.New("genai request failed")
	}
	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return err
	}
	return nil
}
