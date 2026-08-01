package main

import (
	"fmt"
	"log-monitors/internal/cache"
	"log-monitors/internal/config"
	"log-monitors/internal/db"
	"log-monitors/internal/handler"
	"log-monitors/internal/hub"
	"log-monitors/internal/ingest"
)

func main() {
	cfg := config.LoadConfig()
	dbConn := db.Connect(cfg.DB)
	defer dbConn.Close()
	redisConn := cache.Connect(cfg.Redis)
	defer redisConn.Close()
	fmt.Println("db and redis connected successfully")
	h := hub.NewHub()
	ing := ingest.NewIngest(dbConn, redisConn, h)
	ing.Start(10)
	server := handler.NewServer(cfg, dbConn, redisConn, h, ing)
	r := handler.SetupRouter(server)
	r.Run(":3000")
}
