ALTER TABLE channel_notifications
  ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3,
  ADD COLUMN IF NOT EXISTS signature TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_dispatch_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS last_error TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS dead_letter_reason TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_channel_notifications_dispatch
  ON channel_notifications (status, attempts, created_at DESC);
