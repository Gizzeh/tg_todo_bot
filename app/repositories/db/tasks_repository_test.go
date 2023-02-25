package db

import (
	"github.com/pkg/errors"
	"reflect"
	"testing"
	"tg_todo_bot/app/models"
	"tg_todo_bot/app/repositories/types"
	"tg_todo_bot/config"
	"tg_todo_bot/kernel/db"
	zap_logger "tg_todo_bot/kernel/logger"
	"time"
)

func getTaskRepository() (*TasksRepository, error) {
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

	taskRepository := NewTasksRepository(logger, pgInstance)

	return taskRepository, nil
}

func getTaskModelForCreation() models.Task {
	tomorrow := time.Now().Add(time.Hour * 24)
	return models.Task{
		Title:       "Test task title",
		Description: "Test task description",
		Deadline:    &tomorrow,
		Done:        false,
	}
}

func TestCreate(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.deleteByID(createdTask.ID)

	if createdTask.ID == 0 {
		t.Fatal("task wasn't created but there are no errors")
	}
}

func TestFindByID(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.deleteByID(createdTask.ID)

	findByIdResult, err := repository.FindByID(createdTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(createdTask, findByIdResult) {
		t.Fatal("models not equal")
	}
}

func TestUpdate(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.deleteByID(createdTask.ID)

	createdTask.Title = "Updated task title"
	createdTask.Done = true
	err = repository.Update(createdTask)

	updatedTask, err := repository.FindByID(createdTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(createdTask, updatedTask) {
		t.Fatal("models not equal")
	}
}

func TestDeleteByID(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(createdTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = repository.deleteByID(createdTask.ID)

	_, err = repository.FindByID(createdTask.ID)
	if err == nil {
		t.Fatal("model still exists after delete operation")
	} else {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	}
}

func TestGetAllActive(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	allActive, err := repository.GetAllActive()
	allActiveBeforeCreationCount := len(allActive)

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.deleteByID(createdTask.ID)

	allActive, err = repository.GetAllActive()

	if len(allActive) == allActiveBeforeCreationCount || len(allActive) == 0 {
		t.Fatal("errors occurred during GetAllActive logic")
	}
}

func TestSearchActiveByDeadline(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.deleteByID(createdTask.ID)

	from := createdTask.Deadline.Add(-time.Hour)
	to := createdTask.Deadline.Add(time.Hour)

	searchResult, err := repository.SearchActiveByDeadline(&from, nil)
	if err != nil {
		t.Fatal(err)
	}

	var containCreatedTask = func(createdTask models.Task, searchResult []models.Task) bool {
		found := false
		for _, task := range searchResult {
			if task.ID == createdTask.ID {
				found = true
				break
			}
		}
		return found
	}

	if !containCreatedTask(createdTask, searchResult) {
		t.Fatal("model not found")
	}

	searchResult, err = repository.SearchActiveByDeadline(nil, &to)
	if err != nil {
		t.Fatal(err)
	}
	if !containCreatedTask(createdTask, searchResult) {
		t.Fatal("model not found")
	}

	searchResult, err = repository.SearchActiveByDeadline(&from, &to)
	if err != nil {
		t.Fatal(err)
	}
	if !containCreatedTask(createdTask, searchResult) {
		t.Fatal("model not found")
	}

	_, err = repository.SearchActiveByDeadline(nil, nil)
	if err == nil {
		t.Fatal("from and to is nil but there are no errors")
	}
}

func TestDeleteCompleted(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}

	createdTask.Done = true

	err = repository.Update(createdTask)
	if err != nil {
		t.Fatal(err)
	}

	err = repository.DeleteCompleted()
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(createdTask.ID)
	if err == nil {
		repository.deleteByID(createdTask.ID)
		t.Fatal("model still exists after deletion")
	} else {
		if !errors.Is(err, types.ErrNotFound) {
			repository.deleteByID(createdTask.ID)
			t.Fatal(err)
		}
	}
}
