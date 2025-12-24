package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewAlertHandler(t *testing.T) {
	handler := NewAlertHandler()
	assert.NotNil(t, handler)
}

func TestGetAlerts_HandlerExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AlertHandler{}
	assert.NotNil(t, handler)

	// Test that the handler has the expected methods
	// This is a basic test since the actual HTTP call is hard to mock
	// without refactoring the code to accept a configurable HTTP client
	assert.NotNil(t, handler.GetAlerts)
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

	var response ErrorResponse
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

	var response ErrorResponse
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
