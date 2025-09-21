CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_delivered BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS events_recipient_id_is_delivered_idx ON events (recipient_id, is_delivered);

