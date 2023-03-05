package db

import (
	"github.com/pkg/errors"
	"reflect"
	"testing"
	"tg_todo_bot/config"
	"tg_todo_bot/kernel/db"
	zap_logger "tg_todo_bot/kernel/logger"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
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

func getTaskModelForCreation() (models.Task, error) {
	user, err := createUserForTest()
	if err != nil {
		return models.Task{}, err
	}

	tomorrow := time.Now().Add(time.Hour * 24)
	return models.Task{
		Title:       "Test task title",
		Description: "Test task description",
		Datetime:    &tomorrow,
		Done:        false,
		UserID:      user.ID,
		User:        &user,
	}, nil
}

func createTaskForTest() (models.Task, error) {
	taskModel, err := getTaskModelForCreation()
	if err != nil {
		return models.Task{}, err
	}

	repository, err := getTaskRepository()
	if err != nil {
		return models.Task{}, err
	}

	task, err := repository.Create(taskModel)
	if err != nil {
		return models.Task{}, err
	}

	return task, nil
}

func deleteTaskAfterTest(task models.Task) error {
	repository, err := getTaskRepository()
	if err != nil {
		return err
	}

	err = repository.DeleteByID(task.ID)
	if err != nil {
		return err
	}

	err = deleteUserAfterTest(*task.User)
	if err != nil {
		return err
	}

	return nil
}

func TestCreateTask(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	if taskModel.ID == 0 {
		t.Fatal("task wasn't created but there are no errors")
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindTaskByID(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	findByIdResult, err := repository.FindByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	findByIdResult.User = taskModel.User

	if reflect.DeepEqual(taskModel, findByIdResult) {
		t.Fatal("models not equal")
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateTask(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	taskModel.Title = "Updated task title"
	taskModel.Done = true
	err = repository.Update(taskModel)

	updatedTask, err := repository.FindByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}
	updatedTask.User = taskModel.User

	if taskModel.ID != updatedTask.ID {
		t.Fatal("models not equal")
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteTaskByID(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(taskModel.ID)
	if err == nil {
		t.Fatal("model still exists after delete operation")
	} else {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAllActiveTasks(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	allActive, err := repository.GetAllActiveForUser(taskModel.UserID)
	allActiveBeforeCreationCount := len(allActive)

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	allActive, err = repository.GetAllActiveForUser(taskModel.UserID)

	if len(allActive) == allActiveBeforeCreationCount || len(allActive) == 0 {
		t.Fatal("errors occurred during GetAllActive logic")
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchActiveTasksByDatetime(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	from := taskModel.Datetime.Add(-time.Hour)
	to := taskModel.Datetime.Add(time.Hour)

	searchResult, err := repository.SearchActiveByDatetimeForUser(&from, nil, taskModel.UserID)
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

	if !containCreatedTask(taskModel, searchResult) {
		t.Fatal("model not found")
	}

	searchResult, err = repository.SearchActiveByDatetimeForUser(nil, &to, taskModel.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if !containCreatedTask(taskModel, searchResult) {
		t.Fatal("model not found")
	}

	searchResult, err = repository.SearchActiveByDatetimeForUser(&from, &to, taskModel.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if !containCreatedTask(taskModel, searchResult) {
		t.Fatal("model not found")
	}

	_, err = repository.SearchActiveByDatetimeForUser(nil, nil, taskModel.UserID)
	if err == nil {
		t.Fatal("from and to is nil but there are no errors")
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCompletedTasks(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err = repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	taskModel.Done = true

	err = repository.Update(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	err = repository.DeleteCompleted()
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByID(taskModel.ID)
	if err == nil {
		repository.DeleteByID(taskModel.ID)
		t.Fatal("model still exists after deletion")
	} else {
		if !errors.Is(err, types.ErrNotFound) {
			repository.DeleteByID(taskModel.ID)
			t.Fatal(err)
		}
	}

	err = repository.DeleteByID(taskModel.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindTasksByIDs(t *testing.T) {
	repository, err := getTaskRepository()
	if err != nil {
		t.Fatal(err)
	}

	taskModel, err := getTaskModelForCreation()
	if err != nil {
		t.Fatal(err)
	}

	task1, err := repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	task2, err := repository.Create(taskModel)
	if err != nil {
		t.Fatal(err)
	}

	ids := []int64{task1.ID, task2.ID}

	tasks, err := repository.FindByIDs(ids)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 2 {
		t.Fatal("the method should return two tasks, but did not return")
	}

	for _, task := range tasks {
		if task.ID != task1.ID && task.ID != task2.ID {
			t.Fatal("unknown task")
		}
	}

	err = repository.DeleteByID(task1.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = repository.DeleteByID(task2.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserAfterTest(*taskModel.User)
	if err != nil {
		t.Fatal(err)
	}
}
