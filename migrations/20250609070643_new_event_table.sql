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
    status VARCHAR(100)
);

CREATE TABLE pipelines (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER,
    event_id INTEGER,
    user_id VARCHAR(100),
    template_id INTEGER,
    sending_counter INTEGER DEFAULT 0,
    status VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pipeline_templates (
    id SERIAL PRIMARY KEY,
    event_name TEXT,
    execute_type TEXT DEFAULT 'default',
    conditions JSONB,
    query JSONB,
    exit_pipeline_name TEXT,
    next_pipeline_id INTEGER,
    execute_delay INTEGER,
    is_active BOOLEAN
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events
DROP TABLE IF EXISTS pipelines
-- +goose StatementEnd
