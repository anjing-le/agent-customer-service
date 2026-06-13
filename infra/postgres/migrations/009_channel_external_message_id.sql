ALTER TABLE channel_inbound_events
  ADD COLUMN IF NOT EXISTS external_message_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_inbound_events_external_message
  ON channel_inbound_events (channel, external_message_id)
  WHERE external_message_id <> '';
