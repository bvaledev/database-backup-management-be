
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE datasources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    database VARCHAR NOT NULL,
    host VARCHAR NOT NULL,
    port INTEGER NOT NULL,
    ssl_mode VARCHAR NOT NULL,
    username VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    cron_expr TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL
);

CREATE TABLE backups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    datasource_id UUID NOT NULL REFERENCES datasources(id) ON DELETE CASCADE,
    trigger VARCHAR NOT NULL CHECK (trigger IN ('manual', 'cron')),
    status VARCHAR NOT NULL CHECK (status IN ('initialized', 'completed', 'failed')),
    file_name VARCHAR,
    file_size BIGINT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    restored_at TIMESTAMP
);
