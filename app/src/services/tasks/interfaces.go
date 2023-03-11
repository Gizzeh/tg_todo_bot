package tasks

import (
	"tg_todo_bot/src/models"
	"time"
)

type TasksRepositoryI interface {
	Create(task models.Task) (models.Task, error)
	SearchActiveByDatetimeForUser(from, to *time.Time, userID int64) ([]models.Task, error)
	GetAllActiveForUser(userID int64) ([]models.Task, error)
	Update(model models.Task) error
	DeleteByID(ID int64) error
	DeleteCompleted() error
	FindByID(ID int64) (models.Task, error)
	GetActiveTasksWithoutDatetimeForUser(userID int64) ([]models.Task, error)
}

type NotificationsRepositoryI interface {
	FindByTasksIDs(tasksIDs []int64) (map[int64]models.Notification, error)
}
