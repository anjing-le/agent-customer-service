CREATE TABLE IF NOT EXISTS channel_failure_events (
  id BIGSERIAL PRIMARY KEY,
  channel TEXT NOT NULL,
  code TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  external_conversation_id TEXT NOT NULL DEFAULT '',
  external_message_id TEXT NOT NULL DEFAULT '',
  origin TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_channel_failure_events_created_at
  ON channel_failure_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_channel_failure_events_channel_code
  ON channel_failure_events (channel, code);
