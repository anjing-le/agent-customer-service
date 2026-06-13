CREATE TABLE IF NOT EXISTS channel_inbound_events (
  replay_key TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  external_conversation_id TEXT NOT NULL,
  payload_timestamp TEXT NOT NULL,
  signature TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_channel_inbound_events_channel
  ON channel_inbound_events (channel, received_at DESC);

CREATE INDEX IF NOT EXISTS idx_channel_inbound_events_external
  ON channel_inbound_events (external_conversation_id, received_at DESC);
