package db

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
	"tg_todo_bot/src/models"
	"tg_todo_bot/src/repositories/types"
	"time"
)

//TODO: Добавить метод поиска задач пользователя

type TasksRepository struct {
	logger     *zap.SugaredLogger
	dbInstance *pgx.ConnPool
}

func NewTasksRepository(
	logger *zap.SugaredLogger,
	dbInstance *pgx.ConnPool,
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
				"deadline":    task.Deadline,
				"done":        task.Done,
				"user_id":     task.UserID,
				"created_at":  now,
			},
		).
		Returning("id")

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "CreateTask"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> Create -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return models.Task{}, err
	}

	row := repository.dbInstance.QueryRow(preparedStatementName, args...)

	err = row.Scan(&task.ID)
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
			goqu.C("deadline"),
			goqu.C("done"),
			goqu.C("user_id"),
			goqu.C("created_at"),
		)
}

func (repository *TasksRepository) SearchActiveByDeadlineForUser(from, to *time.Time, userID int64) ([]models.Task, error) {
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
			goqu.C("deadline").Asc(),
			goqu.C("title").Asc(),
		).
		Where(
			goqu.C("done").IsFalse(),
			goqu.C("user_id").Eq(userID),
		)

	if from != nil {
		query = query.Where(
			goqu.C("deadline").Gte(*from),
		)
	}
	if to != nil {
		query = query.Where(
			goqu.C("deadline").Lte(*to),
		)
	}

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "SearchActiveTasksByDeadline"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	rows, err := repository.dbInstance.Query(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Query(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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
			&task.Deadline,
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

	err = repository.dbInstance.Deallocate(preparedStatementName)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Deallocate(preparedStatementName)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName,
		)
		return []models.Task{}, err
	}

	return tasks, nil
}

func (repository *TasksRepository) GetAllActiveForUser(userID int64) ([]models.Task, error) {
	query := repository.selectAllCols().
		Order(
			goqu.C("deadline").Asc(),
			goqu.C("title").Asc(),
		).
		Where(
			goqu.C("done").IsFalse(),
			goqu.C("user_id").Eq(userID),
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "GetAllActiveTasks"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	rows, err := repository.dbInstance.Query(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> SearchByDeadline -> repository.dbInstance.Query(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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
			&task.Deadline,
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
				"deadline":    model.Deadline,
				"done":        model.Done,
				"user_id":     model.UserID,
			},
		)

	sql, args, _ := query.Prepared(true).ToSQL()

	preparedStatementName := "UpdateTask"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> Update -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> Update -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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

	preparedStatementName := "DeleteTaskByIDTasks"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> deleteByID -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> deleteByID -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
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

	preparedStatementName := "DeleteCompletedTasks"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> DeleteCompleted -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	_, err = repository.dbInstance.Exec(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> DeleteCompleted -> repository.dbInstance.Exec(preparedStatementName, args...)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return err
	}

	return nil
}

func (repository *TasksRepository) FindByID(ID int64) (models.Task, error) {
	tasks, err := repository.FindByIDs([]int64{ID})
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByID -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
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

	preparedStatementName := "FindTasksByIDs"
	_, err := repository.dbInstance.Prepare(preparedStatementName, sql)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByIDs -> repository.dbInstance.Prepare(preparedStatementName, sql)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	rows, err := repository.dbInstance.Query(preparedStatementName, args...)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByIDs -> repository.dbInstance.Query(preparedStatementName, args...)`,
			"error", err.Error(),
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
			&task.Deadline,
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

	err = repository.dbInstance.Deallocate(preparedStatementName)
	if err != nil {
		repository.logger.Debugw(
			`Repositories -> DB -> TasksRepository -> FindByIDs -> repository.dbInstance.Deallocate(preparedStatementName)`,
			"error", err.Error(), "preparedStatementName", preparedStatementName, "SQL", sql, "args", args,
		)
		return []models.Task{}, err
	}

	return tasks, nil
}