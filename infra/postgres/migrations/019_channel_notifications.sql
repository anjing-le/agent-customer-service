CREATE TABLE IF NOT EXISTS channel_notifications (
  id TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  severity TEXT NOT NULL,
  target TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'OPEN',
  reason TEXT NOT NULL,
  count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  acked_by TEXT NOT NULL DEFAULT '',
  ack_note TEXT NOT NULL DEFAULT '',
  acked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_channel_notifications_status
  ON channel_notifications (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_channel_notifications_channel
  ON channel_notifications (channel, created_at DESC);
