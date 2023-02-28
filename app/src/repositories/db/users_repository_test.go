package db

import (
	"github.com/pkg/errors"
	"math/rand"
	"testing"
	"tg_todo_bot/config"
	"tg_todo_bot/kernel/db"
	zap_logger "tg_todo_bot/kernel/logger"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
	"time"
)

func getUsersRepository() (*UsersRepository, error) {
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

	usersRepository := NewUsersRepository(logger, pgInstance)
	return usersRepository, nil
}

func generateRandomTelegramID() int64 {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 99999999
	return int64(rand.Intn(max-min+1) + min)
}

func createUserForTest() (models.User, error) {
	repository, err := getUsersRepository()
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		TelegramID: generateRandomTelegramID(),
	}

	user, err = repository.Create(user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func deleteUserAfterTest(user models.User) error {
	repository, err := getUsersRepository()
	if err != nil {
		return err
	}

	err = repository.DeleteByTelegramID(user.TelegramID)
	if err != nil {
		return err
	}

	return nil
}

func TestCreateUser(t *testing.T) {
	repository, err := getUsersRepository()
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		TelegramID: generateRandomTelegramID(),
	}

	user, err = repository.Create(user)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.DeleteByTelegramID(user.TelegramID)

	findUserResult, err := repository.FindByTelegramID(user.TelegramID)
	if err != nil {
		t.Fatal(err)
	}

	if findUserResult.ID != user.ID {
		t.Fatal("models are not equal")
	}

	user, err = repository.Create(user)
	if err == nil {
		t.Fatal("user already exist by there are no errors")
	} else {
		if !errors.Is(err, types.ErrAlreadyExist) {
			t.Fatal(err)
		}
	}
}

func TestFindUserByID(t *testing.T) {
	repository, err := getUsersRepository()
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		TelegramID: generateRandomTelegramID(),
	}

	_, err = repository.FindByTelegramID(user.TelegramID)
	if err == nil {
		t.Fatal("user not exist but there are no errors")
	} else {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	}

	user, err = repository.Create(user)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.DeleteByTelegramID(user.TelegramID)

	findResult, err := repository.FindByTelegramID(user.TelegramID)
	if err != nil {
		t.Fatal(err)
	}

	if findResult.ID != user.ID {
		t.Fatal("models are not equal")
	}
}

func TestDeleteUserByTelegramID(t *testing.T) {
	repository, err := getUsersRepository()
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		TelegramID: generateRandomTelegramID(),
	}

	user, err = repository.Create(user)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByTelegramID(user.TelegramID)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = repository.DeleteByTelegramID(user.TelegramID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.FindByTelegramID(user.TelegramID)
	if err == nil {
		t.Fatal("user still exist after deletion")
	} else {
		if !errors.Is(err, types.ErrNotFound) {
			t.Fatal(err)
		}
	}
}
