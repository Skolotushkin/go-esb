package adapter

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-esb/internal/models"
)

// SOAPAdapter реализует SOAP протокол
type SOAPAdapter struct {
	client *http.Client
}

// NewSOAPAdapter создает новый SOAP адаптер
func NewSOAPAdapter() *SOAPAdapter {
	return &SOAPAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SOAPEnvelope представляет SOAP обертку
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  *SOAPHeader
	Body    SOAPBody
}

// SOAPHeader представляет SOAP заголовок
type SOAPHeader struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`
	Content []byte   `xml:",innerxml"`
}

// SOAPBody представляет SOAP тело
type SOAPBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Content []byte   `xml:",innerxml"`
}

// Send отправляет SOAP запрос (action содержит SOAPAction)
func (s *SOAPAdapter) Send(ctx context.Context, endpoint string, action string, headers map[string]string, body []byte) ([]byte, int, error) {
	soapAction := action
	// Обертка SOAP тела в Envelope
	envelope := SOAPEnvelope{
		Body: SOAPBody{
			Content: body,
		},
	}

	xmlData, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(xmlData))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set SOAP headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if soapAction != "" {
		req.Header.Set("SOAPAction", soapAction)
	}

	// Set additional headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read SOAP response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, fmt.Errorf("SOAP error %d: %s", resp.StatusCode, string(respBody))
	}

	// Извлечение тела из SOAP ответа
	bodyContent, err := s.extractBodyFromSOAPResponse(respBody)
	if err != nil {
		// Если не удалось извлечь, возвращаем весь ответ
		return respBody, resp.StatusCode, nil
	}

	return bodyContent, resp.StatusCode, nil
}

// extractBodyFromSOAPResponse извлекает тело из SOAP ответа
func (s *SOAPAdapter) extractBodyFromSOAPResponse(soapResp []byte) ([]byte, error) {
	var envelope SOAPEnvelope
	if err := xml.Unmarshal(soapResp, &envelope); err != nil {
		return nil, err
	}
	return envelope.Body.Content, nil
}

// Authenticate выполняет аутентификацию для SOAP API
func (s *SOAPAdapter) Authenticate(auth *models.ConnectionAuthentication, endpoint string) (map[string]string, error) {
	headers := make(map[string]string)

	switch auth.Type {
	case models.AuthBasic:
		if auth.Username != "" && auth.Password != "" {
			req, _ := http.NewRequest("GET", endpoint, nil)
			req.SetBasicAuth(auth.Username, auth.Password)
			headers["Authorization"] = req.Header.Get("Authorization")
		}
	}

	return headers, nil
}

