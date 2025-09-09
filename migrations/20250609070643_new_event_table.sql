-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    device_carrier VARCHAR(100),
    device_family VARCHAR(100),
    device_id VARCHAR(100),
    device_type VARCHAR(100),
    display_name VARCHAR(100),
    dma VARCHAR(100),
    event_id INTEGER UNIQUE,
    event_properties JSONB,
    event_time VARCHAR(100),
    event_type VARCHAR(100),
    user_id VARCHAR(100),
    user_properties JSONB,
    uuid VARCHAR(100),
    version_name VARCHAR(100),
    is_sended BOOL DEFAULT FALSE,
    is_planned BOOL DEFAULT FALSE,
    sending_attempts INTEGER DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events
-- +goose StatementEnd
