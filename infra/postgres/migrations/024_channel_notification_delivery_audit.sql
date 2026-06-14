ALTER TABLE channel_notifications
  ADD COLUMN IF NOT EXISTS delivery_audit JSONB NOT NULL DEFAULT '[]'::jsonb;
