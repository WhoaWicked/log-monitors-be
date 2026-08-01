package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log-monitors/internal/hub"
	"log-monitors/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Ingest struct {
	jobs  chan *models.Log
	db    *sqlx.DB
	redis *redis.Client
	hub   *hub.Hub
}

type alertRuleCheck struct {
	ID            string `db:"id"`
	Threshold     int    `db:"threshold"`
	WindowSeconds int    `db:"window_seconds"`
}

func NewIngest(db *sqlx.DB, redis *redis.Client, hub *hub.Hub) *Ingest {
	return &Ingest{
		jobs:  make(chan *models.Log, 100),
		db:    db,
		redis: redis,
		hub:   hub,
	}
}

func (i *Ingest) worker() {
	for job := range i.jobs {
		insertCtx, insertCancel := context.WithTimeout(context.Background(), 3*time.Second)
		var err error
		createQuery := `INSERT INTO logs (timestamp, service, level, message) VALUES($1, $2, $3, $4);`
		for attempt := 1; attempt <= 3; attempt++ {
			_, err = i.db.ExecContext(insertCtx, createQuery, job.Timestamp, job.Service, job.Level, job.Message)
			if err == nil {
				break
			}
			log.Printf("insert attempt %d/3 failed: %v", attempt, err)
			if attempt < 3 {
				time.Sleep(200 * time.Millisecond)
			}
		}
		insertCancel()
		if err != nil {
			log.Printf("insert log PERMANENTLY failed after 3 attempts: %v", err)
			continue
		}
		alertCtx, alertCancel := context.WithTimeout(context.Background(), 3*time.Second)
		getQuery := `SELECT id, threshold, window_seconds
		FROM alert_rules
		WHERE enabled = true AND service = $1 AND level = $2;`
		rule := new(alertRuleCheck)
		err = i.db.GetContext(alertCtx, rule, getQuery, job.Service, job.Level)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("get alert rule failed: %v", err)
			}
			alertCancel()
			continue
		}
		key := fmt.Sprintf("alert:count:%s", rule.ID)
		count, err := i.redis.Incr(alertCtx, key).Result()
		if err != nil {
			log.Printf("redis incr failed: %v", err)
			alertCancel()
			continue
		}
		if count == 1 {
			i.redis.Expire(alertCtx, key, time.Duration(rule.WindowSeconds)*time.Second)
		}
		if count >= int64(rule.Threshold) {
			alertMsg := fmt.Sprintf(`{"type":"alert", "service":"%s", "level":"%s", "count":"%d", "threshold":"%d"}`,
				job.Service, job.Level, count, rule.Threshold)
			log.Printf("ALERT TRIGGERED: %s", alertMsg)
			i.hub.Broadcast([]byte(alertMsg))
		}
		alertCancel()
	}
}

func (i *Ingest) Start(numWorkers int) {
	for range numWorkers {
		go i.worker()
	}
}

func (i *Ingest) Enqueue(job *models.Log) {
	i.jobs <- job
}
