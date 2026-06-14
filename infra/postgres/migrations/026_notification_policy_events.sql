CREATE TABLE IF NOT EXISTS notification_policy_events (
  id TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  action TEXT NOT NULL DEFAULT 'UPDATE',
  actor TEXT NOT NULL DEFAULT 'ops-a',
  before_summary TEXT NOT NULL DEFAULT '',
  after_summary TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_policy_events_channel
  ON notification_policy_events (channel, created_at DESC);
