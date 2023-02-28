DROP INDEX IF EXISTS notifications_index_task_id;
ALTER TABLE notifications
    DROP CONSTRAINT fk_task;

DROP TABLE IF EXISTS notifications;

DROP INDEX IF EXISTS notifications_index_notify_at;