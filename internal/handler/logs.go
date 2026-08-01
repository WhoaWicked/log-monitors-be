package handler

import (
	"log-monitors/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateLogRequest struct {
	Timestamp time.Time `json:"timestamp" binding:"required"`
	Service   string    `json:"service" binding:"required"`
	Level     string    `json:"level" binding:"required"`
	Message   string    `json:"message" binding:"required"`
}

func (s *Server) CreateLog(c *gin.Context) {
	req := new(CreateLogRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	// defer cancel()
	// query := `INSERT INTO logs (timestamp, service, level, message) VALUES($1, $2, $3, $4);`
	// if _, err := s.db.ExecContext(ctx, query, req.Timestamp, req.Service, req.Level, req.Message); err != nil {
	// 	log.Printf("insert log failed: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save log"})
	// 	return
	// }
	log := &models.Log{
		Timestamp: req.Timestamp,
		Service:   req.Service,
		Level:     req.Level,
		Message:   req.Message,
	}
	s.ingest.Enqueue(log)
	c.JSON(http.StatusAccepted, gin.H{"message": "create log success"})
}
