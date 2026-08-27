package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateLog_MissingField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body, _ := json.Marshal(map[string]string{
		"timestamp": "2026-08-02T10:00:00Z",
		"level":     "error",
		"message":   "test",
		// ไม่มี "service" — ตั้งใจให้ bind fail
	})
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	server := &Server{}
	server.CreateLog(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
