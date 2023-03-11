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

type NotificationsRepository struct {
	logger     *zap.SugaredLogger
	dbInstance *pgxpool.Pool
}

func NewNotificationsRepository(
	logger *zap.SugaredLogger,
	dbInstance *pgxpool.Pool,
) *NotificationsRepository {
	return &NotificationsRepository{
		logger:     logger,
		dbInstance: dbInstance,
	}
}

func (repository *NotificationsRepository) Create(notification models.Notification) (models.Notification, error) {
	now := time.Now()
	query := goqu.Dialect("postgres").
		Insert("notifications").
		Rows(
			goqu.Record{
				"task_id":         notification.TaskID,
				"notify_at":       notification.NotifyAt,
				"repeat_interval": notification.RepeatInterval,
				"created_at":      now,
			},
		).
		Returning("id")

	sql, args, _ := query.Prepared(true).ToSQL()

	row := repository.dbInstance.QueryRow(context.Background(), sql, args...)

	err := row.Scan(&notification.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			//Нарушение уникальности
			if pgError.Code == "23505" {
				err = types.ErrAlreadyExist
			}
		}
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> Create -> row.Scan()`,
			"error", err.Error(),
		)
		return models.Notification{}, err
	}

	return notification, nil
}

func (repository *NotificationsRepository) Update(notification models.Notification) error {
	query := goqu.Dialect("postgres").
		Update("notifications").
		Set(
			goqu.Record{
				"task_id":         notification.TaskID,
				"notify_at":       notification.NotifyAt,
				"repeat_interval": notification.RepeatInterval,
			},
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	_, err := repository.dbInstance.Exec(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> Update -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}

func (repository *NotificationsRepository) DeleteByID(ID int64) error {
	query := goqu.Dialect("postgres").
		Delete("notifications").
		Where(
			goqu.C("id").Eq(ID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	_, err := repository.dbInstance.Exec(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> DeleteByID -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}

func (repository *NotificationsRepository) selectAllCols() *goqu.SelectDataset {
	return goqu.Dialect("postgres").
		From("notifications").
		Select(
			goqu.C("id"),
			goqu.C("task_id"),
			goqu.C("notify_at"),
			goqu.C("repeat_interval"),
			goqu.C("created_at"),
		)
}

func (repository *NotificationsRepository) FindByTasksIDs(tasksIds []int64) (map[int64]models.Notification, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("task_id").In(tasksIds),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	rows, err := repository.dbInstance.Query(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> FindByTasksIDs -> repository.dbInstance.Query(sql, args...)`,
			"error", err.Error(), "sql", sql, "args", args,
		)
	}

	tasksNotificationsMap := map[int64]models.Notification{}

	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.ID,
			&notification.TaskID,
			&notification.NotifyAt,
			&notification.RepeatInterval,
			&notification.CreatedAt,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				err = types.ErrNotFound
			}
			repository.logger.Debugw(
				`Repositories -> DB -> NotificationsRepository -> FindByTasksIDs -> rows.Scan()`,
				"error", err.Error(),
			)
			return map[int64]models.Notification{}, err
		}

		tasksNotificationsMap[notification.TaskID] = notification
	}

	return tasksNotificationsMap, nil
}

func (repository *NotificationsRepository) GetUpcoming(upcomingTo time.Time) ([]models.Notification, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("notify_at").Lte(upcomingTo),
		).
		Order(
			goqu.C("notify_at").Asc(),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	rows, err := repository.dbInstance.Query(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetUpcoming -> repository.dbInstance.Query(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return []models.Notification{}, err
	}

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		err = rows.Scan(
			&notification.ID,
			&notification.TaskID,
			&notification.NotifyAt,
			&notification.RepeatInterval,
			&notification.CreatedAt,
		)
		if err != nil {
			repository.logger.Debugw(
				`Repositories -> DB -> NotificationsRepository -> GetUpcoming -> rows.Scan()`,
				"error", err.Error(),
			)
			return []models.Notification{}, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (repository *NotificationsRepository) FindByID(ID int64) (models.Notification, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("id").Eq(ID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	row := repository.dbInstance.QueryRow(context.Background(), sql, args...)

	var notification models.Notification
	err := row.Scan(
		&notification.ID,
		&notification.TaskID,
		&notification.NotifyAt,
		&notification.RepeatInterval,
		&notification.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = types.ErrNotFound
		}
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> FindByID -> row.Scan()`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return models.Notification{}, err
	}

	return notification, nil
}
