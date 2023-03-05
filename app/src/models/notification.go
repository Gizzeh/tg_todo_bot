package models

import "time"

type Notification struct {
	ID             int64
	TaskID         int64
	NotifyAt       time.Time
	RepeatInterval time.Duration
	CreatedAt      time.Time

	Task *Task //relation OneToOne
}
