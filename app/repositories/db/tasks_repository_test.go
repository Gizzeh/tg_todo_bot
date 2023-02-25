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

func TestCRUD(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	activeTasks, err := repository.GetAllActive()
	if err != nil {
		t.Fatal(err)
	}

	activeTasksCount := len(activeTasks)

	testTaskForCreation := getTaskModelForCreation()

	createdTask, err := repository.Create(testTaskForCreation)
	if err != nil {
		t.Fatal(err)
	}

	if createdTask.ID == 0 {
		t.Fatal("task wasn't created but there are no errors")
	}

	activeTasks, err = repository.GetAllActive()
	if err != nil {
		t.Fatal(err)
	}

	if activeTasksCount == len(activeTasks) || len(activeTasks) == 0 {
		t.Fatal("method repository.GetAllActive() wasn't found created task")
	}

	findByIdResult, err := repository.FindByID(createdTask.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(findByIdResult, createdTask) {
		t.Fatal("models not equal")
	}

	now := time.Now()
	activeTasksFromToday, err := repository.SearchActiveByDeadline(&now, nil)
	if err != nil {
		t.Fatal(err)
	}

	haveCreatedTask := false
	for _, task := range activeTasksFromToday {
		if task.ID == createdTask.ID {
			haveCreatedTask = true
			break
		}
	}
	if !haveCreatedTask {
		t.Fatal("can't find created model by deadline")
	}

	createdTask.Done = true
	err = repository.Update(createdTask)
	if err != nil {
		t.Fatal(err)
	}

	findByIdResult, err = repository.FindByID(createdTask.ID)
	if err != nil {
		t.Fatal(err)
	}
	if findByIdResult.Done != false {
		t.Fatal("update model wasn't work")
	}

	err = repository.DeleteCompleted()
	if err != nil {
		t.Fatal(err)
	}

	findByIdResult, err = repository.FindByID(createdTask.ID)
	if err != nil {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	} else {
		t.Fatal("model still exists after delete")
	}
}
