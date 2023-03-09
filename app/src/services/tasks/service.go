package tasks

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	repositories_types "tg_todo_bot/src/repositories/types"
	"tg_todo_bot/src/services/tasks/types"
	services_types "tg_todo_bot/src/services/types"
)

type Service struct {
	logger                  *zap.SugaredLogger
	tasksRepository         TasksRepositoryI
	notificationsRepository NotificationsRepositoryI
}

func NewService(
	logger *zap.SugaredLogger,
	tasksRepository TasksRepositoryI,
	notificationsRepository NotificationsRepositoryI,
) *Service {
	return &Service{
		logger:                  logger,
		tasksRepository:         tasksRepository,
		notificationsRepository: notificationsRepository,
	}
}

func (service *Service) Create(params types.CreateParams) error {
	service.logger.Info("Services -> Tasks -> Create")

	err := validateCreateParams(params)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> Create -> validateCreateParams(params)",
			"error", err.Error(), "params", params,
		)
		return err
	}

	taskModel := models.Task{
		Title:       params.Title,
		Description: params.Description,
		Datetime:    params.Datetime,
		Done:        false,
		UserID:      params.UserID,
	}
	taskModel, err = service.tasksRepository.Create(taskModel)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> Create -> service.tasksRepository.Create(taskModel)",
			"error", err.Error(), "params", params, "taskModel", taskModel,
		)
		return err
	}

	return nil
}
func (service *Service) Update(params types.UpdateParams) error {
	service.logger.Info("Services -> Tasks -> Update")

	err := validateUpdateParams(params)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> Update -> validateUpdateParams(params)",
			"error", err.Error(), "params", params,
		)
		return err
	}

	task, err := service.tasksRepository.FindByID(params.TaskID)
	if err != nil {
		if errors.Is(err, repositories_types.ErrNotFound) {
			err = services_types.ErrNotFound
		}
		service.logger.Errorw(
			"Services -> Tasks -> Update -> service.tasksRepository.FindByID(taskID)",
			"error", err.Error(), "params", params, "taskID", params.TaskID,
		)
		return err
	}

	if params.UserID.IsSet {
		task.UserID = params.UserID.Value
	}
	if params.Title.IsSet {
		task.Title = params.Title.Value
	}
	if params.Description.IsSet {
		task.Description = params.Description.Value
	}
	if params.Datetime.IsSet {
		task.Datetime = params.Datetime.Value
	}
	task.Done = params.Done

	err = service.tasksRepository.Update(task)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> Update -> service.tasksRepository.Update(task)",
			"error", err.Error(), "task", task,
		)
		return err
	}

	return nil
}
