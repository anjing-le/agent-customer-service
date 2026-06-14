ALTER TABLE channel_notifications
  ADD COLUMN IF NOT EXISTS target_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS backoff_seconds INTEGER NOT NULL DEFAULT 60,
  ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS receipt_status TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS receipt_body TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_channel_notifications_retry
  ON channel_notifications (status, next_retry_at);
