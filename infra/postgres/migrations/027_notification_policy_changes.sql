CREATE TABLE IF NOT EXISTS notification_policy_changes (
  id TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  target_url TEXT NOT NULL,
  secret_ref TEXT NOT NULL,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  backoff_seconds INTEGER NOT NULL DEFAULT 60,
  requested_by TEXT NOT NULL DEFAULT 'ops-a',
  status TEXT NOT NULL DEFAULT 'PENDING',
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  approved_by TEXT NOT NULL DEFAULT '',
  approved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notification_policy_changes_status
  ON notification_policy_changes (status, created_at DESC);
