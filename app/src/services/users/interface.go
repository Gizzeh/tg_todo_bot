package users

import "tg_todo_bot/src/models"

type UsersRepositoryI interface {
	Create(user models.User) (models.User, error)
	FindByTelegramID(telegramID int64) (models.User, error)
	DeleteByTelegramID(telegramID int64) error
}
