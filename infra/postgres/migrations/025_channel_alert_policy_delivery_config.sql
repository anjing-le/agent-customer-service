ALTER TABLE channel_alert_policies
  ADD COLUMN IF NOT EXISTS target_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS secret_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3,
  ADD COLUMN IF NOT EXISTS backoff_seconds INTEGER NOT NULL DEFAULT 60;

UPDATE channel_alert_policies
SET
  target_url = CASE
    WHEN target_url <> '' THEN target_url
    WHEN notify_target LIKE 'http://%' OR notify_target LIKE 'https://%' THEN notify_target
    ELSE 'https://hooks.example.com/anjing/' || notify_target
  END,
  secret_ref = CASE
    WHEN secret_ref <> '' THEN secret_ref
    ELSE 'ANJING_NOTIFICATION_' || upper(replace(replace(replace(notify_target, '-', '_'), '.', '_'), '/', '_')) || '_SECRET'
  END,
  max_attempts = CASE WHEN max_attempts <= 0 THEN 3 ELSE max_attempts END,
  backoff_seconds = CASE WHEN backoff_seconds <= 0 THEN 60 ELSE backoff_seconds END;
