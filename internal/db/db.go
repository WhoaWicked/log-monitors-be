package db

import (
	"fmt"
	"log"
	"log-monitors/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func Connect(cfg config.DBConfig) *sqlx.DB {
	url := fmt.Sprintf(`
	host=%s port=%d user=%s password=%s dbname=%s sslmode=disable`,
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	db, err := sqlx.Connect("pgx", url)
	if err != nil {
		log.Fatalf("connect to db failed: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db failed: %v\n", err)
	}
	return db
}
