CREATE TABLE notifications
(
    id         SERIAL PRIMARY KEY,
    task_id    INTEGER   NOT NULL UNIQUE,
    notify_at  TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_task FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX notifications_index_task_id ON notifications (task_id);

CREATE INDEX notifications_index_notify_at ON notifications (notify_at);