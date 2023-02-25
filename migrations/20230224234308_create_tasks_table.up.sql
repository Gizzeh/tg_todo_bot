CREATE TABLE tasks
(
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    deadline    TIMESTAMP,
    done        BOOLEAN               DEFAULT false,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX tasks_index_deadline ON tasks (deadline);