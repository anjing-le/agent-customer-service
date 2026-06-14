ALTER TABLE notification_policy_changes
  ADD COLUMN IF NOT EXISTS current_target_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS current_secret_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS current_max_attempts INTEGER NOT NULL DEFAULT 3,
  ADD COLUMN IF NOT EXISTS current_backoff_seconds INTEGER NOT NULL DEFAULT 60;

UPDATE notification_policy_changes
SET current_target_url = target_url
WHERE current_target_url = '';

UPDATE notification_policy_changes
SET current_secret_ref = secret_ref
WHERE current_secret_ref = '';

CREATE INDEX IF NOT EXISTS idx_notification_policy_changes_approved_channel
  ON notification_policy_changes (lower(channel), status, approved_at DESC, created_at DESC);
