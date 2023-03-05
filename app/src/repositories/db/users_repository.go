package db

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
	"time"
)

type UsersRepository struct {
	logger     *zap.SugaredLogger
	dbInstance *pgxpool.Pool
}

func NewUsersRepository(
	logger *zap.SugaredLogger,
	dbInstance *pgxpool.Pool,
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
		).
		Returning("id")

	sql, args, _ := query.Prepared(true).ToSQL()

	row := repository.dbInstance.QueryRow(context.Background(), sql, args...)
	err := row.Scan(&user.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			//Нарушение уникальности
			if pgError.Code == "23505" {
				err = types.ErrAlreadyExist
			}
		}
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> Create -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return models.User{}, err
	}

	return user, nil
}

func (repository *UsersRepository) FindByTelegramID(telegramID int64) (models.User, error) {
	query := goqu.Dialect("postgres").
		From("users").
		Select(
			goqu.C("id"),
			goqu.C("telegram_id"),
			goqu.C("created_at"),
		).
		Where(
			goqu.C("telegram_id").Eq(telegramID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	row := repository.dbInstance.QueryRow(context.Background(), sql, args...)

	var user models.User

	err := row.Scan(
		&user.ID,
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

	_, err := repository.dbInstance.Exec(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> UsersRepository -> DeleteByTelegramID -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}
