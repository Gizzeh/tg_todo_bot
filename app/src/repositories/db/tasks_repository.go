package db

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
	"time"
)

//TODO: Добавить метод поиска задач пользователя

type TasksRepository struct {
	logger     *zap.SugaredLogger
	dbInstance *pgxpool.Pool
}

func NewTasksRepository(
	logger *zap.SugaredLogger,
	dbInstance *pgxpool.Pool,
) *TasksRepository {
	return &TasksRepository{
		logger:     logger,
		dbInstance: dbInstance,
	}
}

func (repository *TasksRepository) Create(task models.Task) (models.Task, error) {
	now := time.Now()
	query := goqu.Dialect("postgres").
		Insert("tasks").
		Rows(
			goqu.Record{
				"title":       task.Title,
				"description": task.Description,
				"datetime":    task.Datetime,
				"done":        task.Done,
				"user_id":     task.UserID,
				"created_at":  now,
			},
		).
		Returning("id")

	sql, args, _ := query.Prepared(true).ToSQL()

	row := repository.dbInstance.QueryRow(context.Background(), sql, args...)

	err := row.Scan(&task.ID)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> Create -> row.Scan()`,
			"error", err.Error(),
		)
		return models.Task{}, err
	}

	return task, nil
}

func (repository *TasksRepository) selectAllCols() *goqu.SelectDataset {
	return goqu.Dialect("postgres").
		From("tasks").
		Select(
			goqu.C("id"),
			goqu.C("title"),
			goqu.C("description"),
			goqu.C("datetime"),
			goqu.C("done"),
			goqu.C("user_id"),
			goqu.C("created_at"),
		)
}

func (repository *TasksRepository) SearchActiveByDatetimeForUser(from, to *time.Time, userID int64) ([]models.Task, error) {
	if from == nil && to == nil {
		err := fmt.Errorf(`"from" and "to" are empty`)
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline`,
			"error", err.Error(),
		)
		return []models.Task{}, err
	}

	query := repository.selectAllCols().
		Order(
			goqu.C("datetime").Asc(),
			goqu.C("title").Asc(),
		).
		Where(
			goqu.C("done").IsFalse(),
			goqu.C("user_id").Eq(userID),
		)

	if from != nil {
		query = query.Where(
			goqu.C("datetime").Gte(*from),
		)
	}
	if to != nil {
		query = query.Where(
			goqu.C("datetime").Lte(*to),
		)
	}

	sql, args, _ := query.Prepared(true).ToSQL()

	rows, err := repository.dbInstance.Query(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Query(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	var tasks []models.Task
	for rows.Next() {
		var task models.Task

		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Datetime,
			&task.Done,
			&task.UserID,
			&task.CreatedAt,
		)
		if err != nil {
			repository.logger.Debugw(
				`Repositories -> DB -> TasksRepository -> SearchByDeadline -> rows.Scan()`,
				"error", err.Error(),
			)
			return []models.Task{}, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (repository *TasksRepository) GetAllActiveForUser(userID int64) ([]models.Task, error) {
	query := repository.selectAllCols().
		Order(
			goqu.C("datetime").Asc(),
			goqu.C("title").Asc(),
		).
		Where(
			goqu.C("done").IsFalse(),
			goqu.C("user_id").Eq(userID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	rows, err := repository.dbInstance.Query(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Query(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	var tasks []models.Task
	for rows.Next() {
		var task models.Task

		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Datetime,
			&task.Done,
			&task.UserID,
			&task.CreatedAt,
		)
		if err != nil {
			repository.logger.Debugw(
				`Repositories -> DB -> TasksRepository -> SearchByDeadline -> rows.Scan()`,
				"error", err.Error(),
			)
			return []models.Task{}, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (repository *TasksRepository) Update(model models.Task) error {
	query := goqu.Dialect("postgres").
		Update("tasks").
		Set(
			goqu.Record{
				"title":       model.Title,
				"description": model.Description,
				"datetime":    model.Datetime,
				"done":        model.Done,
				"user_id":     model.UserID,
			},
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	_, err := repository.dbInstance.Exec(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> Update -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}

func (repository *TasksRepository) DeleteByID(ID int64) error {
	query := goqu.Dialect("postgres").
		Delete("tasks").
		Where(
			goqu.C("id").Eq(ID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	_, err := repository.dbInstance.Exec(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> deleteByID -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}

func (repository *TasksRepository) DeleteCompleted() error {
	query := goqu.Dialect("postgres").
		Delete("tasks").
		Where(
			goqu.C("done").IsTrue(),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	_, err := repository.dbInstance.Exec(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> DeleteCompleted -> repository.dbInstance.Exec(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}

func (repository *TasksRepository) FindByID(ID int64) (models.Task, error) {
	tasks, err := repository.FindByIDs([]int64{ID})
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByID -> repository.dbInstance.Prepare(sql, sql)`,
			"error", err.Error(),
		)
		return models.Task{}, err
	}

	if len(tasks) == 0 {
		err = types.ErrNotFound
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByID`,
			"error", err.Error(),
		)
		return models.Task{}, err
	}

	return tasks[0], nil
}

func (repository *TasksRepository) FindByIDs(IDs []int64) ([]models.Task, error) {
	query := repository.selectAllCols().
		Where(
			goqu.C("id").In(IDs),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	rows, err := repository.dbInstance.Query(context.Background(), sql, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByIDs -> repository.dbInstance.Query(sql, args...)`,
			"error", err.Error(), "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Datetime,
			&task.Done,
			&task.UserID,
			&task.CreatedAt,
		)
		if err != nil {
			repository.logger.Debugw(
				`Repositories -> DB -> TasksRepository -> FindByIDs -> row.Scan()`,
				"error", err.Error(),
			)
			return []models.Task{}, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
