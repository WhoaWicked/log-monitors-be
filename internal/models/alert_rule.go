package models

import "time"

type AlertRule struct {
	ID            string    `db:"id" json:"id"`
	Service       string    `db:"service" json:"service"`
	Level         string    `db:"level" json:"level"`
	Threshold     int       `db:"threshold" json:"threshold"`
	WindowSeconds int       `db:"window_seconds" json:"window_seconds"`
	Enabled       bool      `db:"enabled" json:"enabled"`
	CreatedBy     string    `db:"created_by" json:"created_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
