package db

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
	"tg_todo_bot/app/models"
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
				"done":       notification.Done,
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
				"done":      notification.Done,
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
			goqu.C("done"),
			goqu.C("created_at"),
		)

}

func (repository *NotificationsRepository) GetTasksActiveNotifications(tasksIds []int64) (map[int64][]models.Notification, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("task_id").In(tasksIds),
			goqu.C("done").IsFalse(),
		).
		Order(
			goqu.C("notify_at").Asc(),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "GetTasksActiveNotifications"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetTasksActiveNotifications -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return map[int64][]models.Notification{}, err
	}

	rows, err := repository.dbInstance.Query(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetTasksActiveNotifications -> repository.dbInstance.Query(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return map[int64][]models.Notification{}, err
	}

	tasksNotificationsMap := map[int64][]models.Notification{}
	for rows.Next() {
		var notification models.Notification
		err = rows.Scan(
			&notification.ID,
			&notification.TaskID,
			&notification.NotifyAt,
			&notification.Done,
			&notification.CreatedAt,
		)
		if err != nil {
			repository.logger.Debugw(
				`Repositories -> DB -> NotificationsRepository -> GetTasksActiveNotifications -> rows.Scan()`,
				"error", err.Error(),
			)
			return map[int64][]models.Notification{}, err
		}

		notifications, exist := tasksNotificationsMap[notification.TaskID]
		if exist {
			notifications = append(notifications, notification)
		} else {
			notifications = []models.Notification{notification}
		}
		tasksNotificationsMap[notification.TaskID] = notifications
	}

	err = repository.dbInstance.Deallocate(preparedStatementName)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> GetTasksActiveNotifications -> repository.dbInstance.Deallocate(preparedStatementName)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName,
		)
		return map[int64][]models.Notification{}, err
	}

	return tasksNotificationsMap, nil
}

func (repository *NotificationsRepository) GetUpcoming(upcomingTo time.Time) ([]models.Notification, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("done").IsFalse(),
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
			&notification.Done,
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

func (repository *NotificationsRepository) DeleteCompleted() error {
	query := goqu.Dialect("postgres").
		Delete("notifications").
		Where(
			goqu.C("done").IsTrue(),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "DeleteCompletedNotifications"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> DeleteCompleted -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> NotificationsRepository -> DeleteCompleted -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}
