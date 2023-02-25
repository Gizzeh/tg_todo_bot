package models

import "time"

type Task struct {
	ID          int64
	Title       string
	Description string
	Deadline    *time.Time
	Done        bool
	CreatedAt   time.Time
}
