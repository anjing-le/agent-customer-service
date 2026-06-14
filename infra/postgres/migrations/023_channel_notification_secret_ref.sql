ALTER TABLE channel_notifications
  ADD COLUMN IF NOT EXISTS secret_ref TEXT NOT NULL DEFAULT '';

UPDATE channel_notifications
SET secret_ref = 'ANJING_NOTIFICATION_' || upper(replace(replace(replace(target, '-', '_'), '.', '_'), '/', '_')) || '_SECRET'
WHERE secret_ref = '';
