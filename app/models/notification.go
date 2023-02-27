package models

import "time"

type Notification struct {
	ID        int64
	TaskID    int64
	NotifyAt  time.Time
	Done      bool
	CreatedAt time.Time

	Task *Task //relation OneToMany
}
