package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-esb/internal/models"
)

// RESTAdapter реализует REST протокол
type RESTAdapter struct {
	client *http.Client
}

// NewRESTAdapter создает новый REST адаптер
func NewRESTAdapter() *RESTAdapter {
	return &RESTAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send отправляет REST запрос (action содержит HTTP method)
func (r *RESTAdapter) Send(ctx context.Context, endpoint string, action string, headers map[string]string, body []byte) ([]byte, int, error) {
	method := action
	if method == "" {
		method = "POST" // default
	}
	var reqBody io.Reader
	if body != nil && len(body) > 0 {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default content type if body exists
	if body != nil && len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

// Authenticate выполняет аутентификацию для REST API
func (r *RESTAdapter) Authenticate(auth *models.ConnectionAuthentication, endpoint string) (map[string]string, error) {
	headers := make(map[string]string)

	switch auth.Type {
	case models.AuthBasic:
		if auth.Username != "" && auth.Password != "" {
			req, _ := http.NewRequest("GET", endpoint, nil)
			req.SetBasicAuth(auth.Username, auth.Password)
			headers["Authorization"] = req.Header.Get("Authorization")
		}

	case models.AuthBearerToken:
		if auth.Token != "" {
			headers["Authorization"] = "Bearer " + auth.Token
		}
	}

	return headers, nil
}

