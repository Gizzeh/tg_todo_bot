package models

import "time"

type Task struct {
	ID          int64
	Title       string
	Description string
	Deadline    *time.Time
	Done        bool
	UserID      int64
	CreatedAt   time.Time

	User         *User         //relation OneToOne
	Notification *Notification //relation OneToOne
}
