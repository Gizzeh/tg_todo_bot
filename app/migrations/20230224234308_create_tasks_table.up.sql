CREATE TABLE tasks
(
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    datetime    TIMESTAMP,
    done        BOOLEAN      NOT NULL DEFAULT false,
    user_id     INTEGER      NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX tasks_index_datetime ON tasks (datetime);

CREATE INDEX tasks_index_user_id ON tasks (user_id);