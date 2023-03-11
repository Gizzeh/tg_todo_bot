package tasks

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	repositories_types "tg_todo_bot/src/repositories/types"
	"tg_todo_bot/src/services/tasks/types"
	services_types "tg_todo_bot/src/services/types"
	"time"
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

//SearchByDateForUser (params) -> return map[DateWithoutTime][]models.Task
func (service *Service) SearchByDateForUser(params types.SearchByDateForUserParams) (map[time.Time][]models.Task, error) {
	service.logger.Info("Services -> Tasks -> SearchByDateForUser")

	err := validateSearchByDateForUserParams(params)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> SearchByDateForUser -> validateSearchByDateForUserParams(params)",
			"error", err.Error(), "params", params,
		)
		return map[time.Time][]models.Task{}, err
	}

	tasks, err := service.tasksRepository.SearchActiveByDatetimeForUser(params.From, params.To, params.UserID)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> SearchByDateForUser -> service.tasksRepository.SearchActiveByDatetimeForUser(from, to, userID)",
			"error", err.Error(), "from", params.From, "to", params.To, "userID", params.UserID,
		)
		return map[time.Time][]models.Task{}, err
	}

	var tasksIDs []int64
	for _, task := range tasks {
		tasksIDs = append(tasksIDs, task.ID)
	}

	tasksNotificationsMap, err := service.notificationsRepository.FindByTasksIDs(tasksIDs)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> SearchByDateForUser -> service.notificationsRepository.FindByTasksIDs(tasksIDs)",
			"error", err.Error(), "tasksIDs", tasksIDs,
		)
		return map[time.Time][]models.Task{}, err
	}

	dateTasksMap := map[time.Time][]models.Task{}

	for _, task := range tasks {
		notification, exist := tasksNotificationsMap[task.ID]
		if exist {
			task.Notification = &notification
		}

		y, m, d := task.Datetime.Date()
		taskDate := time.Date(y, m, d, 0, 0, 0, 0, nil)
		if tasksByDate, exist := dateTasksMap[taskDate]; exist {
			tasksByDate = append(tasksByDate, task)
		} else {
			dateTasksMap[taskDate] = []models.Task{task}
		}
	}

	return dateTasksMap, nil
}

func (service *Service) GetAllActiveForUser(userID int64) ([]models.Task, error) {
	service.logger.Info("Services -> Tasks -> GetAllActiveForUser")

	tasks, err := service.tasksRepository.GetAllActiveForUser(userID)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> SearchByDateForUser -> service.tasksRepository.GetAllActiveForUser(userID)",
			"error", err.Error(), "userID", userID,
		)
		return []models.Task{}, err
	}

	err = service.setNotifications(tasks)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> SearchByDateForUser -> service.setNotifications(tasks)",
			"error", err.Error(), "tasks", tasks,
		)
		return []models.Task{}, err
	}

	return tasks, nil
}

func (service *Service) setNotifications(tasks []models.Task) error {
	var tasksIDs []int64
	for _, task := range tasks {
		tasksIDs = append(tasksIDs, task.ID)
	}

	tasksNotificationsMap, err := service.notificationsRepository.FindByTasksIDs(tasksIDs)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> getAndSetNotifications -> service.notificationsRepository.FindByTasksIDs(tasksIDs)",
			"error", err.Error(), "tasksIDs", tasksIDs,
		)
		return err
	}

	for i, task := range tasks {
		notification, exist := tasksNotificationsMap[task.ID]
		if exist {
			tasks[i].Notification = &notification
		}
	}

	return nil
}

func (service *Service) GetActiveTasksWithoutDatetimeForUser(userID int64) ([]models.Task, error) {
	service.logger.Info("Services -> Tasks -> GetAllActiveForUser")

	tasks, err := service.tasksRepository.GetActiveTasksWithoutDatetimeForUser(userID)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> GetActiveTasksWithoutDatetimeForUser -> service.tasksRepository.GetActiveTasksWithoutDatetimeForUser(userID)",
			"error", err.Error(), "userID", userID,
		)
		return []models.Task{}, err
	}

	err = service.setNotifications(tasks)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> GetActiveTasksWithoutDatetimeForUser -> service.setNotifications(tasks)",
			"error", err.Error(), "tasks", tasks,
		)
		return []models.Task{}, err
	}

	return tasks, nil
}

func (service *Service) DeleteByID(taskID int64) error {
	service.logger.Info("Services -> Tasks -> DeleteByID")

	err := service.tasksRepository.DeleteByID(taskID)
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> GetActiveTasksWithoutDatetimeForUser -> service.tasksRepository.DeleteByID(taskID)",
			"error", err.Error(), "taskID", taskID,
		)
		return err
	}

	return nil
}

func (service *Service) DeleteCompleted() error {
	service.logger.Info("Services -> Tasks -> DeleteCompleted")

	err := service.tasksRepository.DeleteCompleted()
	if err != nil {
		service.logger.Errorw(
			"Services -> Tasks -> DeleteCompleted -> service.tasksRepository.DeleteCompleted()",
			"error", err.Error(),
		)
		return err
	}

	return nil
}

func (service *Service) FindByID(taskID int64) (models.Task, error) {
	service.logger.Info("Services -> Tasks -> FindByID")

	task, err := service.tasksRepository.FindByID(taskID)
	if err != nil {
		if errors.Is(err, repositories_types.ErrNotFound) {
			err = services_types.ErrNotFound
		}
		service.logger.Errorw(
			"Services -> Tasks -> FindByID -> service.tasksRepository.FindByID(taskID)",
			"error", err.Error(),
		)
		return models.Task{}, err
	}

	return task, nil
}
