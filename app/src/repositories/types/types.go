package types

import "fmt"

var (
	ErrNotFound     = fmt.Errorf("not found")
	ErrAlreadyExist = fmt.Errorf("already exist")
)
