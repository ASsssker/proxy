-- +goose Up
-- +goose StatementBegin
CREATE TYPE statuses AS ENUM('done', 'in process', 'error', 'new');
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    status statuses NOT NULL,
    status_code INT NOT NULL DEFAULT 0,
    headers JSONB NOT NULL DEFAULT '{}'::JSONB,
    body TEXT NOT NULL DEFAULT '',
    content_length INT NOT NULL DEFAULT 0
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
DROP TYPE IF EXISTS statuses;
-- +goose StatementEnd
    