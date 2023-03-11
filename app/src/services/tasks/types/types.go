package types

import "time"

type CreateParams struct {
	Title       string
	Description string
	Datetime    *time.Time
	UserID      int64
}

type UpdateParams struct {
	TaskID int64
	Title  struct {
		Value string
		IsSet bool
	}
	Description struct {
		Value string
		IsSet bool
	}
	Datetime struct {
		Value *time.Time
		IsSet bool
	}
	UserID struct {
		Value int64
		IsSet bool
	}
	Done bool
}

type SearchByDateForUserParams struct {
	From   *time.Time
	To     *time.Time
	UserID int64
}
