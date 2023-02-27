CREATE TABLE tasks
(
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    deadline    TIMESTAMP,
    done        BOOLEAN      NOT NULL DEFAULT false,
    user_id     INTEGER      NOT NULL REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX tasks_index_deadline ON tasks (deadline);

CREATE INDEX tasks_index_user_id ON tasks (user_id);
ALTER TABLE tasks
    ADD CONSTRAINT fk_task_user FOREIGN KEY (user_id) REFERENCES users (id);