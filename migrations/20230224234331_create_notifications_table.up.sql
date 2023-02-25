CREATE TABLE notifications
(
    id         SERIAL PRIMARY KEY,
    task_id    INTEGER   NOT NULL,
    notify_at  TIMESTAMP NOT NULL,
    done       BOOLEAN            DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX notifications_index_task_id ON notifications (task_id);
ALTER TABLE notifications
    ADD CONSTRAINT fk_notification_task FOREIGN KEY (task_id) REFERENCES tasks (id);
CREATE INDEX notifications_index_notify_at ON notifications (notify_at);