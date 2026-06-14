ALTER TABLE notification_policy_changes
  ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

UPDATE notification_policy_changes
SET expires_at = created_at + interval '24 hours'
WHERE expires_at IS NULL;

ALTER TABLE notification_policy_changes
  ALTER COLUMN expires_at SET NOT NULL,
  ALTER COLUMN expires_at SET DEFAULT (now() + interval '24 hours');

CREATE INDEX IF NOT EXISTS idx_notification_policy_changes_expiry
  ON notification_policy_changes (status, expires_at);
