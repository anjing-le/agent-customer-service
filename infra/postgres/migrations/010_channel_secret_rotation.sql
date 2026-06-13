ALTER TABLE channel_integrations
  ADD COLUMN IF NOT EXISTS next_secret_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS rotates_at TIMESTAMPTZ;

UPDATE channel_integrations
SET
  next_secret_ref = CASE channel
    WHEN 'Web' THEN 'ANJING_CHANNEL_WEB_NEXT_SECRET'
    WHEN 'WeChat' THEN 'ANJING_CHANNEL_WECHAT_NEXT_SECRET'
    WHEN 'App' THEN 'ANJING_CHANNEL_APP_NEXT_SECRET'
    WHEN 'Marketplace' THEN 'ANJING_CHANNEL_MARKETPLACE_NEXT_SECRET'
    ELSE next_secret_ref
  END,
  rotation_hint = CASE
    WHEN rotation_hint = '' THEN 'active and next secret refs can overlap during webhook rotation'
    ELSE rotation_hint
  END,
  updated_at = now()
WHERE next_secret_ref = '';
