package db

import (
	"github.com/pkg/errors"
	"testing"
	"tg_todo_bot/config"
	"tg_todo_bot/kernel/db"
	zap_logger "tg_todo_bot/kernel/logger"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
	"time"
)

func getNotificationRepository() (*NotificationsRepository, error) {
	logger := zap_logger.InitLogger()

	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	pg := db.NewPG(
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Database,
		conf.Database.User,
		conf.Database.Password,
	)
	pgInstance, err := pg.OpenPool()
	if err != nil {
		return nil, err
	}

	notificationRepository := NewNotificationsRepository(logger, pgInstance)

	return notificationRepository, nil
}

func getNotificationModelForCreation() (models.Notification, error) {
	task, err := createTaskForTest()
	if err != nil {
		return models.Notification{}, err
	}

	afterAnHour := time.Now().Add(time.Hour)
	return models.Notification{
		TaskID:         task.ID,
		NotifyAt:       afterAnHour,
		RepeatInterval: time.Hour,
		Task:           &task,
	}, nil
}

func TestCreateNotification(t *testing.T) {
	repository, err := getNotificationRepository()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err := getNotificationModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err = repository.Create(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	if notificationModel.ID == 0 {
		t.Fatal("errors occurred during model creation")
	}

	err = repository.DeleteByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteTaskAfterTest(*notificationModel.Task)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateNotification(t *testing.T) {
	repository, err := getNotificationRepository()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err := getNotificationModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err = repository.Create(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	notificationModel.NotifyAt.Add(time.Hour)
	err = repository.Update(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	findResult, err := repository.FindByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	if findResult.NotifyAt != findResult.NotifyAt {
		t.Fatal("models not equal")
	}

	err = repository.DeleteByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteTaskAfterTest(*notificationModel.Task)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteNotificationByID(t *testing.T) {
	repository, err := getNotificationRepository()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err := getNotificationModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err = repository.Create(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = repository.DeleteByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(notificationModel.ID)
	if err != nil {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	} else {
		t.Fatal("model still exist after deletion")
	}

	err = deleteTaskAfterTest(*notificationModel.Task)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTaskNotification(t *testing.T) {
	repository, err := getNotificationRepository()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err := getNotificationModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err = repository.Create(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	taskNotification, err := repository.FindByTaskID(notificationModel.TaskID)
	if err != nil {
		t.Fatal(err)
	}

	taskNotification.Task = notificationModel.Task
	if taskNotification.ID != notificationModel.ID {
		t.Fatal("models not equal")
	}

	err = repository.DeleteByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteTaskAfterTest(*notificationModel.Task)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUpcomingNotifications(t *testing.T) {
	repository, err := getNotificationRepository()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err := getNotificationModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err = repository.Create(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	upcomingNotifications, err := repository.GetUpcoming(time.Now())
	if err != nil {
		t.Fatal(err)
	}

	haveNotification := false
	for _, notification := range upcomingNotifications {
		if notification.ID == notificationModel.ID {
			haveNotification = true
			break
		}
	}
	if haveNotification {
		t.Fatal("error in logic")
	}

	upcomingNotifications, err = repository.GetUpcoming(time.Now().Add(time.Hour * 2))
	if err != nil {
		t.Fatal(err)
	}

	for _, notification := range upcomingNotifications {
		if notification.ID == notificationModel.ID {
			haveNotification = true
			break
		}
	}
	if !haveNotification {
		t.Fatal("model not found")
	}

	err = repository.DeleteByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteTaskAfterTest(*notificationModel.Task)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindNotificationByID(t *testing.T) {
	repository, err := getNotificationRepository()
	if err != nil {
		t.Fatal(err)
	}

	notificationModel, err := getNotificationModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(notificationModel.ID)
	if err != nil {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	} else {
		t.Fatal("must return error")
	}

	notificationModel, err = repository.Create(notificationModel)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = repository.DeleteByID(notificationModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteTaskAfterTest(*notificationModel.Task)
	if err != nil {
		t.Fatal(err)
	}
}
