package models

import "time"

type AlertRule struct {
	ID            string    `db:"id"`
	Service       string    `db:"service"`
	Level         string    `db:"level"`
	Threshold     int       `db:"threshold"`
	WindowSeconds int       `db:"window_seconds"`
	Enabled       bool      `db:"enabled"`
	CreatedBy     string    `db:"created_by"`
	CreatedAt     time.Time `db:"created_at"`
}
