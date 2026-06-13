ALTER TABLE channel_integrations
  ADD COLUMN IF NOT EXISTS allowed_origins TEXT[] NOT NULL DEFAULT '{}';

UPDATE channel_integrations
SET
  allowed_origins = CASE channel
    WHEN 'Web' THEN ARRAY['https://console.example.com']
    WHEN 'WeChat' THEN ARRAY['https://wechat.example.com']
    WHEN 'App' THEN ARRAY['app://agent-customer-service']
    WHEN 'Marketplace' THEN ARRAY['https://marketplace.example.com']
    ELSE allowed_origins
  END,
  updated_at = now()
WHERE cardinality(allowed_origins) = 0;
