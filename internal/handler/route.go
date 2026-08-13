package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(s *Server) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	r.GET("/health", s.HealthCheck)
	r.POST("/login", s.Login)
	protected := r.Group("/")
	protected.Use(s.JwtAuth())
	{
		protected.GET("/logs", s.ListLogs)
		protected.POST("/logs", s.CreateLog)
		protected.POST("/alert-rules", s.CreateAlertRule)
		protected.GET("/alert-rules", s.ListAlertRules)
	}
	r.GET("/ws", s.WebsocketHandler)
	return r
}
