CREATE TABLE users
(
    telegram_id INTEGER   NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX users_index_telegram_id ON users (telegram_id);