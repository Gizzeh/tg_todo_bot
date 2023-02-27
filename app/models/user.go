package models

import "time"

type User struct {
	TelegramID int64
	CreatedAt  time.Time
}
