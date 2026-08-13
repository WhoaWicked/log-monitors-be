package handler

import (
	"log-monitors/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		result, err := auth.VerifyToken(token, []byte(s.cfg.JWT.Secret))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("user_id", result.UserId)
		c.Set("email", result.Email)
		c.Next()
	}
}
