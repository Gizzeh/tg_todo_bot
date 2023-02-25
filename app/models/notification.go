package models

import "time"

type Notification struct {
	ID        int64
	TaskID    int64
	Task      *Task //relation
	NotifyAt  time.Time
	Done      bool
	CreatedAt time.Time
}
