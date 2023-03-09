package users

import (
	"fmt"
	"tg_todo_bot/src/services/users/types"
)

func validateCreateParams(params types.CreateParams) error {
	if params.TelegramID == 0 {
		err := fmt.Errorf("TelegramID is required filed")
		return err
	}
	return nil
}
