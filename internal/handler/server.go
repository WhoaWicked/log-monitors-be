package handler

import (
	"log-monitors/internal/config"
	"log-monitors/internal/hub"
	"log-monitors/internal/ingest"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	db     *sqlx.DB
	redis  *redis.Client
	hub    *hub.Hub
	cfg    *config.Config
	ingest *ingest.Ingest
}

func (s *Server) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func NewServer(cfg *config.Config, db *sqlx.DB, redis *redis.Client, hub *hub.Hub, ingest *ingest.Ingest) *Server {
	return &Server{
		cfg:    cfg,
		db:     db,
		redis:  redis,
		hub:    hub,
		ingest: ingest,
	}
}
