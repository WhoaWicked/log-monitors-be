package models

import "time"

type Log struct {
	ID        string    `db:"id"`
	Timestamp time.Time `db:"timestamp"`
	Service   string    `db:"service"`
	Level     string    `db:"level"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}
