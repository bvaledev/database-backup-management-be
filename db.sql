
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE schedule_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cron_expr TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL,
);

INSERT INTO scheduled_tasks (id, cron_expr, description, enabled) VALUES
(uuid_generate_v4(),'*/1 * * * *', 'Executar a cada 1 minuto', true),
(uuid_generate_v4(),'0 0 * * *', 'Executar todo dia à meia-noite', true);

CREATE TABLE datasources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    database VARCHAR NOT NULL,
    host VARCHAR NOT NULL,
    port INTEGER NOT NULL,
    ssl_mode VARCHAR NOT NULL,
    username VARCHAR NOT NULL,
    password VARCHAR NOT NULL
    cron_expr TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL,
);
