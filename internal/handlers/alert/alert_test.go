package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"taheri24.ir/graph1/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockHTTPClient implements HTTPClient interface for testing
type MockHTTPClient struct {
	Response *http.Response
	Error    error
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.Response, m.Error
}

func TestNewAlertHandler(t *testing.T) {
	handler := NewAlertHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.httpClient)
	assert.Equal(t, "http://prometheus:9090/api/v1/alerts", handler.prometheusURL)
}

func TestNewAlertHandlerWithDeps(t *testing.T) {
	mockClient := &MockHTTPClient{}
	customURL := "http://custom-prometheus:9090/api/v1/alerts"
	handler := NewAlertHandlerWithDeps(mockClient, customURL)
	assert.NotNil(t, handler)
	assert.Equal(t, mockClient, handler.httpClient)
	assert.Equal(t, customURL, handler.prometheusURL)
}

func TestGetAlerts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock response
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
			"status": "success",
			"data": {
				"alerts": [
					{
						"labels": {"alertname": "TestAlert"},
						"annotations": {"summary": "Test alert"},
						"state": "firing",
						"activeAt": "2023-12-24T10:00:00Z",
						"value": "1"
					}
				]
			}
		}`)),
		Header: make(http.Header),
	}

	mockClient := &MockHTTPClient{Response: mockResponse}
	handler := NewAlertHandlerWithDeps(mockClient, "http://test-prometheus/api/v1/alerts")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{Method: "GET"}

	handler.GetAlerts(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PrometheusAlertResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Len(t, response.Data.Alerts, 1)
	assert.Equal(t, "TestAlert", response.Data.Alerts[0].Labels["alertname"])
}

func TestGetAlerts_HTTPError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockHTTPClient{Error: assert.AnError}
	handler := NewAlertHandlerWithDeps(mockClient, "http://test-prometheus/api/v1/alerts")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{Method: "GET"}

	handler.GetAlerts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to query Prometheus", response.Error)
}

func TestGetAlerts_ReadError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create response with body that will fail to read
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       &errorReader{err: assert.AnError},
		Header:     make(http.Header),
	}

	mockClient := &MockHTTPClient{Response: mockResponse}
	handler := NewAlertHandlerWithDeps(mockClient, "http://test-prometheus/api/v1/alerts")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{Method: "GET"}

	handler.GetAlerts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to read Prometheus response", response.Error)
}

func TestGetAlerts_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`invalid json`)),
		Header:     make(http.Header),
	}

	mockClient := &MockHTTPClient{Response: mockResponse}
	handler := NewAlertHandlerWithDeps(mockClient, "http://test-prometheus/api/v1/alerts")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{Method: "GET"}

	handler.GetAlerts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to parse Prometheus response", response.Error)
}

// errorReader is a helper for testing read errors
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReader) Close() error {
	return nil
}

// TestPrometheusAlertResponse_JSON tests the JSON structure
func TestPrometheusAlertResponse_JSON(t *testing.T) {
	response := PrometheusAlertResponse{
		Status: "success",
		Data: struct {
			Alerts []Alert `json:"alerts"`
		}{
			Alerts: []Alert{
				{
					Labels: map[string]string{
						"alertname": "TestAlert",
						"severity":  "critical",
					},
					Annotations: map[string]string{
						"summary": "Test alert summary",
					},
					State:    "firing",
					ActiveAt: "2023-12-24T10:00:00Z",
					Value:    "1",
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"status":"success"`)
	assert.Contains(t, string(data), `"alertname":"TestAlert"`)

	// Test JSON unmarshaling
	var unmarshaled PrometheusAlertResponse
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "success", unmarshaled.Status)
	assert.Len(t, unmarshaled.Data.Alerts, 1)
	assert.Equal(t, "TestAlert", unmarshaled.Data.Alerts[0].Labels["alertname"])
}

// TestAlert_JSON tests the Alert struct JSON marshaling
func TestAlert_JSON(t *testing.T) {
	alert := Alert{
		Labels: map[string]string{
			"alertname": "CPUAlert",
		},
		Annotations: map[string]string{
			"description": "High CPU usage",
		},
		State:    "firing",
		ActiveAt: "2023-12-24T10:00:00Z",
		Value:    "95.5",
	}

	// Test marshaling
	data, err := json.Marshal(alert)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"alertname":"CPUAlert"`)

	// Test unmarshaling
	var unmarshaled Alert
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "CPUAlert", unmarshaled.Labels["alertname"])
	assert.Equal(t, "firing", unmarshaled.State)
}

func TestFireAlert_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AlertHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request body
	reqBody := FireAlertRequest{
		AlertName: "TestAlert",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Set up the request
	c.Request = &http.Request{
		Method: "POST",
		Body:   io.NopCloser(strings.NewReader(string(jsonBody))),
	}

	// Call the handler
	handler.FireAlert(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Alert triggered successfully", response["message"])
	assert.Equal(t, "TestAlert", response["alert_name"])
}

func TestFireAlert_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AlertHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up invalid JSON request
	c.Request = &http.Request{
		Method: "POST",
		Body:   io.NopCloser(strings.NewReader(`{"alert_name":`)), // Invalid JSON
	}

	// Call the handler
	handler.FireAlert(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Error)
}

func TestResetAlert_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AlertHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request body
	reqBody := FireAlertRequest{
		AlertName: "TestAlert",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Set up the request
	c.Request = &http.Request{
		Method: "POST",
		Body:   io.NopCloser(strings.NewReader(string(jsonBody))),
	}

	// Call the handler
	handler.ResetAlert(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Alert reset successfully", response["message"])
	assert.Equal(t, "TestAlert", response["alert_name"])
}

func TestResetAlert_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AlertHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up invalid JSON request
	c.Request = &http.Request{
		Method: "POST",
		Body:   io.NopCloser(strings.NewReader(`{"alert_name":`)), // Invalid JSON
	}

	// Call the handler
	handler.ResetAlert(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Error)
}

func TestAlertRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectError bool
		alertName   string
	}{
		{
			name:        "valid request",
			requestBody: `{"alert_name":"TestAlert"}`,
			expectError: false,
			alertName:   "TestAlert",
		},
		{
			name:        "missing alert_name",
			requestBody: `{"other_field":"value"}`,
			expectError: false, // JSON unmarshaling succeeds, field gets zero value
			alertName:   "",    // zero value for string
		},
		{
			name:        "empty alert_name",
			requestBody: `{"alert_name":""}`,
			expectError: false, // JSON unmarshaling succeeds
			alertName:   "",    // empty string
		},
		{
			name:        "invalid json",
			requestBody: `{"alert_name":}`,
			expectError: true, // Invalid JSON syntax
			alertName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req FireAlertRequest
			err := json.Unmarshal([]byte(tt.requestBody), &req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.alertName, req.AlertName)
			}
		})
	}
}

// TestGinBindingValidation tests the actual Gin binding validation that happens in handlers
func TestGinBindingValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		requestBody  string
		expectStatus int
	}{
		{
			name:         "valid request",
			requestBody:  `{"alert_name":"TestAlert"}`,
			expectStatus: http.StatusOK,
		},
		{
			name:         "missing alert_name",
			requestBody:  `{"other_field":"value"}`,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "empty alert_name",
			requestBody:  `{"alert_name":""}`,
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_fire", func(t *testing.T) {
			handler := &AlertHandler{}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = &http.Request{
				Method: "POST",
				Body:   io.NopCloser(strings.NewReader(tt.requestBody)),
			}

			handler.FireAlert(c)
			assert.Equal(t, tt.expectStatus, w.Code)
		})

		t.Run(tt.name+"_reset", func(t *testing.T) {
			handler := &AlertHandler{}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = &http.Request{
				Method: "POST",
				Body:   io.NopCloser(strings.NewReader(tt.requestBody)),
			}

			handler.ResetAlert(c)
			assert.Equal(t, tt.expectStatus, w.Code)
		})
	}
}
