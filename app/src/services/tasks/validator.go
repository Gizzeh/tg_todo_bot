package tasks

import (
	"fmt"
	"strings"
	"tg_todo_bot/src/services/tasks/types"
)

func validateCreateParams(params types.CreateParams) error {
	var emptyRequiredFields []string

	if params.UserID == 0 {
		emptyRequiredFields = append(emptyRequiredFields, "UserID")
	}

	if params.Title == "" {
		emptyRequiredFields = append(emptyRequiredFields, "Title")
	}

	if len(emptyRequiredFields) > 0 {
		err := fmt.Errorf("some required fields are empty: [%s]", strings.Join(emptyRequiredFields, ", "))
		return err
	}

	return nil
}

func validateUpdateParams(params types.UpdateParams) error {
	if params.TaskID == 0 {
		err := fmt.Errorf("TaskID is required field")
		return err
	}

	var emptyFields []string

	if params.Title.IsSet && params.Title.Value == "" {
		emptyFields = append(emptyFields, "title")
	}

	if params.UserID.IsSet && params.UserID.Value == 0 {
		emptyFields = append(emptyFields, "userID")
	}

	if len(emptyFields) > 0 {
		err := fmt.Errorf("some fields are empty: [%s]", strings.Join(emptyFields, ", "))
		return err
	}

	return nil
}

func validateSearchByDateForUserParams(params types.SearchByDateForUserParams) error {
	if params.UserID == 0 {
		err := fmt.Errorf("UserID can't be empty")
		return err
	}

	if params.From == nil && params.To == nil {
		err := fmt.Errorf("at least one of ['from', 'to'] must be not null")
		return err
	}

	return nil
}
