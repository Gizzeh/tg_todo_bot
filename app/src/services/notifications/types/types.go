package types

import "time"

type CreateParams struct {
	TaskID         int64
	NotifyAt       time.Time
	RepeatInterval time.Duration
}

type UpdateParams struct {
	NotificationID int64
	NotifyAt       struct {
		Value time.Time
		IsSet bool
	}
	RepeatInterval struct {
		Value time.Duration
		IsSet bool
	}
}
