package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log-monitors/internal/models"
	"net/http"
	"strconv"
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

// func (s *Server) ListLogs(c *gin.Context) {
// 	service := c.Query("service")
// 	level := c.Query("level")
// 	limitStr := c.DefaultQuery("limit", "20")
// 	offsetStr := c.DefaultQuery("offset", "0")
// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit <= 0 {
// 		limit = 20
// 	}
// 	offset, err := strconv.Atoi(offsetStr)
// 	if err != nil || offset < 0 {
// 		offset = 0
// 	}
// 	query := `SELECT id, timestamp, service, level, message, created_at FROM logs WHERE 1=1`
// 	countQuery := `SELECT COUNT(*) as total FROM logs WHERE 1=1`
// 	conditions := make([]string, 0)
// 	filterValues := make([]any, 0)
// 	countValues := make([]any, 0)
// 	if service != "" {
// 		filterValues = append(filterValues, service)
// 		countValues = append(countValues, service)
// 		conditions = append(conditions, fmt.Sprintf(" AND service = $%d", len(filterValues)))
// 	}
// 	if level != "" {
// 		filterValues = append(filterValues, level)
// 		countValues = append(countValues, level)
// 		conditions = append(conditions, fmt.Sprintf(" AND level = $%d", len(filterValues)))
// 	}
// 	if len(conditions) > 0 {
// 		query += strings.Join(conditions, "")
// 		countQuery += strings.Join(conditions, "")
// 	}
// 	filterValues = append(filterValues, limit)
// 	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", len(filterValues))
// 	filterValues = append(filterValues, offset)
// 	query += fmt.Sprintf(" OFFSET $%d", len(filterValues))
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
// 	defer cancel()
// 	listLogs := make([]*models.Log, 0)
// 	if err := s.db.SelectContext(ctx, &listLogs, query, filterValues...); err != nil {
// 		log.Printf("get list logs failed: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "get list logs failed"})
// 		return
// 	}
// 	var total int
// 	if err := s.db.GetContext(ctx, &total, countQuery, countValues...); err != nil {
// 		log.Printf("total list logs failed: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "total list logs failed"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"total": total, "data": listLogs})
// }

func (s *Server) ListLogs(c *gin.Context) {
	service := c.Query("service")
	level := c.Query("level")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	conditions := `WHERE 1=1`
	args := make([]any, 0)
	if service != "" {
		args = append(args, service)
		conditions += fmt.Sprintf(" AND service = $%d", len(args))
	}
	if level != "" {
		args = append(args, level)
		conditions += fmt.Sprintf(" AND level = $%d", len(args))
	}
	selectQuery := fmt.Sprintf(`SELECT id, timestamp, service, level, message, created_at FROM logs %s
	ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`, conditions, len(args)+1, len(args)+2)
	countQuery := fmt.Sprintf(`SELECT COUNT(*) as total FROM logs %s`, conditions)
	listLogs := make([]*models.Log, 0)
	selectArgs := make([]any, len(args), len(args)+2)
	copy(selectArgs, args)
	selectArgs = append(selectArgs, limit, offset)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		log.Printf("begin transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "begin transaction failed"})
		return
	}
	defer tx.Rollback()
	if err := tx.SelectContext(ctx, &listLogs, selectQuery, selectArgs...); err != nil {
		log.Printf("get list logs failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get list logs failed"})
		return
	}
	var total int
	if err := tx.GetContext(ctx, &total, countQuery, args...); err != nil {
		log.Printf("total list logs failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "total list logs failed"})
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("commit transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit transaction failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "data": listLogs})
}
