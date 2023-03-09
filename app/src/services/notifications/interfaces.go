package notifications

import (
	"tg_todo_bot/src/models"
	"time"
)

type NotificationsRepositoryI interface {
	Create(notification models.Notification) (models.Notification, error)
	Update(notification models.Notification) error
	DeleteByID(ID int64) error
	GetUpcoming(upcomingTo time.Time) ([]models.Notification, error)
	FindByID(ID int64) (models.Notification, error)
}
