package db

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"math/rand"
	"reflect"
	"testing"
	"tg_todo_bot/app/models"
	"tg_todo_bot/app/repositories/types"
	"tg_todo_bot/config"
	"tg_todo_bot/kernel/db"
	zap_logger "tg_todo_bot/kernel/logger"
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
	buf := make([]byte, 8)
	rand.Read(buf)
	return int64(binary.LittleEndian.Uint64(buf))
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

	if !reflect.DeepEqual(findUserResult, user) {
		t.Fatal("models not equal")
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

	if !reflect.DeepEqual(findResult, user) {
		t.Fatal("models not equal")
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
