package handler

import (
	"context"
	"log"
	"log-monitors/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateAlertRuleRequest struct {
	Service       string `json:"service" binding:"required"`
	Level         string `json:"level" binding:"required"`
	Threshold     int    `json:"threshold" binding:"required"`
	WindowSeconds int    `json:"window_seconds" binding:"required"`
}

func (s *Server) CreateAlertRule(c *gin.Context) {
	req := new(CreateAlertRuleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*5)
	defer cancel()
	query := `INSERT INTO alert_rules (service, level, threshold, window_seconds, created_by)
	VALUES($1, $2, $3, $4, $5);`
	if _, err := s.db.ExecContext(ctx, query,
		req.Service, req.Level, req.Threshold, req.WindowSeconds, userIdStr); err != nil {
		log.Printf("insert alert rule failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create alert rule failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "create alert rule success"})
}

func (s *Server) ListAlertRules(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*5)
	defer cancel()
	rules := make([]*models.AlertRule, 0)
	query := `SELECT id, service, level, threshold, window_seconds, enabled, created_by, created_at FROM alert_rules;`
	if err := s.db.SelectContext(ctx, &rules, query); err != nil {
		log.Printf("list alert rules failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get list rules failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}
