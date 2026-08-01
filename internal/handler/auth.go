package handler

import (
	"context"
	"log-monitors/internal/auth"
	"log-monitors/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) Login(c *gin.Context) {
	req := new(LoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	user := new(models.User)
	query := "SELECT * FROM users WHERE email = $1;"
	if err := s.db.GetContext(ctx, user, query, req.Email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	token, err := auth.GenerateToken(
		req.Email,
		user.ID,
		[]byte(s.cfg.JWT.Secret),
		s.cfg.JWT.ExpiryHour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate token failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": token})
}
