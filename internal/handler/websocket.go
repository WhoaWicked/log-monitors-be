package handler

import (
	"log"
	"log-monitors/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) WebsocketHandler(c *gin.Context) {
	token := c.Query("token")
	_, err := auth.VerifyToken(token, []byte(s.cfg.JWT.Secret))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	s.hub.Register(conn)
	defer s.hub.Unregister(conn)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
