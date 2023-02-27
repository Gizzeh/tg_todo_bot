package db

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/app/models"
	"tg_todo_bot/app/repositories/types"
	"time"
)

type UsersRepository struct {
	logger     *zap.SugaredLogger
	dbInstance *pgx.ConnPool
}

func NewUsersRepository(
	logger *zap.SugaredLogger,
	dbInstance *pgx.ConnPool,
) *UsersRepository {
	return &UsersRepository{
		logger:     logger,
		dbInstance: dbInstance,
	}
}

func (repository *UsersRepository) Create(user models.User) (models.User, error) {
	now := time.Now()
	query := goqu.Dialect("postgres").
		Insert("users").
		Rows(
			goqu.Record{
				"telegram_id": user.TelegramID,
				"created_at":  now,
			},
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "CreateUser"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> Create -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.User{}, err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		var pgError pgx.PgError
		if errors.As(err, &pgError) {
			//Нарушение уникальности
			if pgError.Code == "23505" {
				err = types.ErrAlreadyExist
			}
		}
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> Create -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.User{}, err
	}

	return user, nil
}

func (repository *UsersRepository) FindByTelegramID(telegramID int64) (models.User, error) {
	query := goqu.Dialect("postgres").
		From("users").
		Select(
			goqu.C("telegram_id"),
			goqu.C("created_at"),
		).
		Where(
			goqu.C("telegram_id").Eq(telegramID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "FindUserByTelegramID"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> FindByTelegramID -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.User{}, err
	}

	row := repository.dbInstance.QueryRow(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> FindByTelegramID -> repository.dbInstance.QueryRow(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.User{}, err
	}

	var user models.User

	err = row.Scan(
		&user.TelegramID,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = types.ErrNotFound
		}
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> FindByTelegramID -> row.Scan()`,
			"error", err.Error(),
		)
		return models.User{}, err
	}

	return user, nil
}

func (repository *UsersRepository) DeleteByTelegramID(telegramID int64) error {
	query := goqu.Dialect("postgres").
		Delete("users").
		Where(
			goqu.C("telegram_id").Eq(telegramID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "DeleteUserByTelegramID"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> DeleteByTelegramID -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> DeleteByTelegramID -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}
