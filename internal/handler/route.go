package handler

import "github.com/gin-gonic/gin"

func SetupRouter(s *Server) *gin.Engine {
	r := gin.Default()
	r.GET("/health", s.HealthCheck)
	r.POST("/login", s.Login)
	protected := r.Group("/")
	protected.Use(s.JwtAuth())
	{
		protected.POST("/logs", s.CreateLog)
		protected.POST("/alert-rules", s.CreateAlertRule)
		protected.GET("/alert-rules", s.ListAlertRules)
	}
	r.GET("/ws", s.WebsocketHandler)
	return r
}
