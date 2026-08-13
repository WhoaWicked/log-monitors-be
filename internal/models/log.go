package models

import "time"

type Log struct {
	ID        string    `db:"id" json:"id"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	Service   string    `db:"service" json:"service"`
	Level     string    `db:"level" json:"level"`
	Message   string    `db:"message" json:"message"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
