DROP INDEX IF EXISTS tasks_index_user_id;
ALTER TABLE tasks
    DROP CONSTRAINT fk_task_user;

DROP TABLE IF EXISTS tasks;

DROP INDEX IF EXISTS tasks_index_deadline;