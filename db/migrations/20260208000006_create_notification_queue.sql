-- migrate:up
CREATE TABLE notification_queue (
    id BIGSERIAL PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'processed', 'failed')),
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX idx_notification_queue_actionable ON notification_queue (status, next_retry_at) WHERE status IN ('pending', 'failed');
CREATE INDEX idx_notification_queue_workspace_id ON notification_queue (workspace_id);
CREATE INDEX idx_notification_queue_created_at ON notification_queue (created_at);

-- migrate:down
DROP TABLE notification_queue;
