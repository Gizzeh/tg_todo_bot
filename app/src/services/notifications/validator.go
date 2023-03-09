package notifications

import (
	"fmt"
	"strings"
	"tg_todo_bot/src/services/notifications/types"
)

func validateCreateParams(params types.CreateParams) error {
	var emptyRequiredFields []string

	if params.NotifyAt.IsZero() {
		emptyRequiredFields = append(emptyRequiredFields, "NotifyAt")
	}

	if params.TaskID == 0 {
		emptyRequiredFields = append(emptyRequiredFields, "TaskID")
	}

	if len(emptyRequiredFields) > 0 {
		err := fmt.Errorf("some required fields are empty: [%s]", strings.Join(emptyRequiredFields, ", "))
		return err
	}

	return nil
}

func validateUpdateParams(params types.UpdateParams) error {
	if params.NotificationID == 0 {
		err := fmt.Errorf("NotifcationID is required field")
		return err
	}

	if !params.NotifyAt.IsSet && !params.RepeatInterval.IsSet {
		err := fmt.Errorf("for update you must set at least one field")
		return err
	}

	if params.NotifyAt.IsSet && params.NotifyAt.Value.IsZero() {
		err := fmt.Errorf("field 'notifyAt' can't be empty")
		return err
	}

	return nil
}
