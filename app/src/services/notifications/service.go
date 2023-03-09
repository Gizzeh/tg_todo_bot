package notifications

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	repositories_types "tg_todo_bot/src/repositories/types"
	"tg_todo_bot/src/services/notifications/types"
	services_types "tg_todo_bot/src/services/types"
	"time"
)

type Service struct {
	logger                  *zap.SugaredLogger
	notificationsRepository NotificationsRepositoryI
}

func NewService(
	logger *zap.SugaredLogger,
	notificationsRepository NotificationsRepositoryI,
) *Service {
	return &Service{
		logger:                  logger,
		notificationsRepository: notificationsRepository,
	}
}

func (service *Service) Create(params types.CreateParams) error {
	service.logger.Info("Services -> Notifications -> Create")

	err := validateCreateParams(params)
	if err != nil {
		service.logger.Errorw(
			"Services -> Notifications -> Create -> validateCreateParams(params)",
			"error", err.Error(), "params", params,
		)
		return err
	}

	if params.RepeatInterval == 0 {
		params.RepeatInterval = time.Hour
	}

	if params.RepeatInterval < time.Minute {
		params.RepeatInterval = time.Minute
	}

	notificationModel := models.Notification{
		TaskID:         params.TaskID,
		NotifyAt:       params.NotifyAt,
		RepeatInterval: params.RepeatInterval,
	}

	notificationModel, err = service.notificationsRepository.Create(notificationModel)
	if err != nil {
		service.logger.Errorw(
			"Services -> Notifications -> Create -> service.notificationsRepository.Create(notificationModel)",
			"error", err.Error(), "notificationModel", notificationModel,
		)
		return err
	}

	return nil
}

func (service *Service) Update(params types.UpdateParams) error {
	service.logger.Info("Services -> Notifications -> Update")

	err := validateUpdateParams(params)
	if err != nil {
		service.logger.Errorw(
			"Services -> Notifications -> Update -> validateUpdateParams(params)",
			"error", err.Error(), "params", params,
		)
		return err
	}

	notificationModel, err := service.notificationsRepository.FindByID(params.NotificationID)
	if err != nil {
		if errors.Is(err, repositories_types.ErrNotFound) {
			err = services_types.ErrNotFound
		}
		service.logger.Errorw(
			"Services -> Notifications -> Update -> service.notificationsRepository.FindByID(notificationID)",
			"error", err.Error(), "notificationID", params.NotificationID, "params", params,
		)
		return err
	}

	if params.NotifyAt.IsSet {
		notificationModel.NotifyAt = params.NotifyAt.Value
	}

	if params.RepeatInterval.IsSet {
		if params.RepeatInterval.Value == 0 {
			params.RepeatInterval.Value = time.Hour
		}

		if params.RepeatInterval.Value < time.Minute {
			params.RepeatInterval.Value = time.Minute
		}

		notificationModel.RepeatInterval = params.RepeatInterval.Value
	}

	err = service.notificationsRepository.Update(notificationModel)
	if err != nil {
		service.logger.Errorw(
			"Services -> Notifications -> Update -> service.notificationsRepository.Update(notificationModel)",
			"error", err.Error(), "notificationModel", notificationModel, "params", params,
		)
		return err
	}

	return nil
}

func (service *Service) DeleteByID(notificationID int64) error {
	service.logger.Info("Services -> Notifications -> DeleteByID")

	err := service.notificationsRepository.DeleteByID(notificationID)
	if err != nil {
		service.logger.Errorw(
			"Services -> Notifications -> DeleteByID -> service.notificationsRepository.DeleteByID(notificationID)",
			"error", err.Error(), "notificationID", notificationID,
		)
		return err
	}

	return nil
}

func (service *Service) GetUpcoming(upcomingTo time.Time) ([]models.Notification, error) {
	service.logger.Info("Services -> Notifications -> GetUpcoming")

	if upcomingTo.IsZero() {
		upcomingTo = time.Now().Add(time.Minute)
	}

	notifications, err := service.notificationsRepository.GetUpcoming(upcomingTo)
	if err != nil {
		service.logger.Errorw(
			"Services -> Notifications -> GetUpcoming -> service.notificationsRepository.GetUpcoming(upcomingTo)",
			"error", err.Error(), "upcomingTo", upcomingTo,
		)
		return []models.Notification{}, err
	}

	return notifications, nil
}
