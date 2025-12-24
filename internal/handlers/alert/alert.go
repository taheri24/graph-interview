package alert

import (
	"encoding/json"
	"io"
	"net/http"

	"taheri24.ir/graph1/internal/dto"
	"taheri24.ir/graph1/internal/middleware"

	"github.com/gin-gonic/gin"
)

type AlertHandler struct{}

func NewAlertHandler() *AlertHandler {
	return &AlertHandler{}
}

// PrometheusAlertResponse represents the response from Prometheus /api/v1/alerts
type PrometheusAlertResponse struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []Alert `json:"alerts"`
	} `json:"data"`
}

// Alert represents a Prometheus alert
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    string            `json:"activeAt"`
	Value       string            `json:"value"`
}

// GetAlerts handles GET /alerts
// @Summary Get current alerts from Prometheus
// @Description Retrieve active alerts from the Prometheus monitoring system
// @Tags alerts
// @Accept json
// @Produce json
// @Success 200 {object} PrometheusAlertResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/alerts [get]
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	// Query Prometheus API
	resp, err := http.Get("http://prometheus:9090/api/v1/alerts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to query Prometheus"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to read Prometheus response"))
		return
	}

	var promResp PrometheusAlertResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to parse Prometheus response"))
		return
	}

	c.JSON(http.StatusOK, promResp)
}

// FireAlertRequest represents the request body for firing an alert
type FireAlertRequest struct {
	AlertName string `json:"alert_name" binding:"required"`
}

// FireAlert handles POST /alerts/fire
// @Summary Manually fire an alert
// @Description Manually trigger an alert by setting the alert trigger metric
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert body FireAlertRequest true "Alert to fire"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/alerts/fire [post]
func (h *AlertHandler) FireAlert(c *gin.Context) {
	var req FireAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErr(err))
		return
	}

	// Trigger the alert by setting the metric
	middleware.TriggerAlert(req.AlertName)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Alert triggered successfully",
		"alert_name": req.AlertName,
	})
}

// ResetAlert handles POST /alerts/reset
// @Summary Reset an alert
// @Description Reset an alert by clearing the alert trigger metric
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert body FireAlertRequest true "Alert to reset"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/alerts/reset [post]
func (h *AlertHandler) ResetAlert(c *gin.Context) {
	var req FireAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErr(err))
		return
	}

	// Reset the alert by clearing the metric
	middleware.ResetAlert(req.AlertName)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Alert reset successfully",
		"alert_name": req.AlertName,
	})
}
