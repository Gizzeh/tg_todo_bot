package users

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	repositories_types "tg_todo_bot/src/repositories/types"
	services_types "tg_todo_bot/src/services/types"
	"tg_todo_bot/src/services/users/types"
)

type Service struct {
	logger          *zap.SugaredLogger
	usersRepository UsersRepositoryI
}

func NewService(
	logger *zap.SugaredLogger,
	usersRepository UsersRepositoryI,
) *Service {
	return &Service{
		logger:          logger,
		usersRepository: usersRepository,
	}
}

func (service *Service) Create(params types.CreateParams) error {
	service.logger.Info("Services -> Users -> Create")

	err := validateCreateParams(params)
	if err != nil {
		service.logger.Errorw(
			"Services -> Users -> Create -> validateCreateParams(params)",
			"error", err.Error(), "params", params,
		)
		return err
	}

	userModes := models.User{
		TelegramID: params.TelegramID,
	}

	userModes, err = service.usersRepository.Create(userModes)
	if err != nil {
		if errors.Is(err, repositories_types.ErrAlreadyExist) {
			return nil
		}
		service.logger.Errorw(
			"Services -> Users -> Create -> service.usersRepository.Create(userModes)",
			"error", err.Error(), "userModes", userModes, "params", params,
		)
		return err
	}

	return nil
}

func (service *Service) FindByTelegramID(telegramID int64) (models.User, error) {
	service.logger.Info("Services -> Users -> FindByTelegramID")

	userModel, err := service.usersRepository.FindByTelegramID(telegramID)
	if err != nil {
		if errors.Is(err, repositories_types.ErrNotFound) {
			err = services_types.ErrNotFound
		}
		service.logger.Errorw(
			"Services -> Users -> FindByTelegramID -> service.usersRepository.FindByTelegramID(telegramID)",
			"error", err.Error(), "telegramID", telegramID,
		)
		return models.User{}, err
	}

	return userModel, nil
}

func (service *Service) DeleteByTelegramID(telegramID int64) error {
	service.logger.Info("Services -> Users -> DeleteByTelegramID")

	err := service.usersRepository.DeleteByTelegramID(telegramID)
	if err != nil {
		service.logger.Errorw(
			"Services -> Users -> DeleteByTelegramID -> service.usersRepository.DeleteByTelegramID(telegramID)",
			"error", err.Error(), "telegramID", telegramID,
		)
		return err
	}

	return nil
}
