-- Inbox Start Scenario table for idempotent message processing
CREATE TABLE IF NOT EXISTS inbox_start_scenario (
    outbox_uuid UUID NOT NULL PRIMARY KEY,
    camera_id INTEGER NOT NULL,
    scenario_uuid UUID NOT NULL,
    url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'received' CHECK (status IN ('received', 'in_process', 'processed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

COMMENT ON TABLE inbox_start_scenario IS 'Inbox pattern table for idempotent message processing in SAGA (start scenario events)';
COMMENT ON COLUMN inbox_start_scenario.outbox_uuid IS 'Unique identifier from the outbox message (serves as primary key for idempotency)';
COMMENT ON COLUMN inbox_start_scenario.camera_id IS 'ID of the camera associated with the scenario';
COMMENT ON COLUMN inbox_start_scenario.scenario_uuid IS 'UUID of the scenario being started';
COMMENT ON COLUMN inbox_start_scenario.url IS 'URL associated with the scenario';
COMMENT ON COLUMN inbox_start_scenario.status IS 'Processing status: received, in_process, processed';
COMMENT ON COLUMN inbox_start_scenario.created_at IS 'Timestamp when the message was first received';
COMMENT ON COLUMN inbox_start_scenario.updated_at IS 'Timestamp when the message status was last updated';

-- Worker table
CREATE TABLE IF NOT EXISTS worker (
    id SERIAL PRIMARY KEY,
    camera_id INTEGER NOT NULL UNIQUE,
    scenario_uuid UUID NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'stopped', 'failed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Node table (runner instances)
CREATE TABLE IF NOT EXISTS node (
    id TEXT PRIMARY KEY,
    addr TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Node Worker table (assigns workers to nodes)
CREATE TABLE IF NOT EXISTS node_worker (
    id SERIAL PRIMARY KEY,
    node_id TEXT NOT NULL REFERENCES node(id) ON DELETE CASCADE,
    worker_id INTEGER NOT NULL REFERENCES worker(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(node_id, worker_id)
);
