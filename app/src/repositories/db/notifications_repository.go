package db

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
	"time"
)

type NotificationsRepository struct {
	logger     *zap.SugaredLogger
	dbInstance *pgx.ConnPool
}

func NewNotificationsRepository(
	logger *zap.SugaredLogger,
	dbInstance *pgx.ConnPool,
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
				"task_id":    notification.TaskID,
				"notify_at":  notification.NotifyAt,
				"created_at": now,
			},
		).
		Returning("id")

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "CreateNotification"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> Create -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.Notification{}, err
	}

	row := repository.dbInstance.QueryRow(preparedStatementName, args...)

	err = row.Scan(&notification.ID)
	if err != nil {
		var pgError pgx.PgError
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
				"task_id":   notification.TaskID,
				"notify_at": notification.NotifyAt,
			},
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "UpdateNotification"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> Update -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> Update -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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

	preparedStatementName := "DeleteNotificationByID"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> DeleteByID -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> DeleteByID -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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
			goqu.C("created_at"),
		)

}

func (repository *NotificationsRepository) GetTaskNotification(taskId int64) (models.Notification, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("task_id").Eq(taskId),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "GetTasksActiveNotifications"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetTasksActiveNotifications -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.Notification{}, err
	}

	row := repository.dbInstance.QueryRow(preparedStatementName, args...)

	var notification models.Notification

	err = row.Scan(
		&notification.ID,
		&notification.TaskID,
		&notification.NotifyAt,
		&notification.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = types.ErrNotFound
		}
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetTasksActiveNotifications -> rows.Scan()`,
			"error", err.Error(),
		)
		return models.Notification{}, err
	}

	return notification, nil
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

	preparedStatementName := "GetUpcomingNotifications"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetUpcoming -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return []models.Notification{}, err
	}

	rows, err := repository.dbInstance.Query(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetUpcoming -> repository.dbInstance.Query(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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

	preparedStatementName := "FindNotificationByID"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> FindByID -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.Notification{}, err
	}

	row := repository.dbInstance.QueryRow(preparedStatementName, args...)

	var notification models.Notification
	err = row.Scan(
		&notification.ID,
		&notification.TaskID,
		&notification.NotifyAt,
		&notification.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = types.ErrNotFound
		}
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> FindByID -> row.Scan()`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.Notification{}, err
	}

	return notification, nil
}
