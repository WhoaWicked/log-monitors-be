package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateAlertRule_MissingField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body, _ := json.Marshal(map[string]any{
		"service":        "payment-service",
		"level":          "error",
		"window_seconds": 60,
		// ไม่มี "threshold" — ตั้งใจให้ bind fail
	})
	req := httptest.NewRequest(http.MethodPost, "/alert-rules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	server := &Server{}
	server.CreateAlertRule(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
