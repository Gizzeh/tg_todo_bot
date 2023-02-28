package models

import "time"

type Notification struct {
	ID        int64
	TaskID    int64
	NotifyAt  time.Time
	CreatedAt time.Time

	Task *Task //relation OneToOne
}
