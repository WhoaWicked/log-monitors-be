package main

import (
	"context"
	"fmt"
	"log"
	"log-monitors/internal/cache"
	"log-monitors/internal/config"
	"log-monitors/internal/db"
	"log-monitors/internal/handler"
	"log-monitors/internal/hub"
	"log-monitors/internal/ingest"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Println("Server started on :3000")
	<-ctx.Done()
	log.Println("shutdown signal received, shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
