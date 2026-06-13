ALTER TABLE channel_integrations
  ADD COLUMN IF NOT EXISTS rate_limit_per_minute INTEGER NOT NULL DEFAULT 60;

UPDATE channel_integrations
SET
  rate_limit_per_minute = CASE channel
    WHEN 'Web' THEN 120
    WHEN 'WeChat' THEN 60
    WHEN 'App' THEN 90
    WHEN 'Marketplace' THEN 45
    ELSE rate_limit_per_minute
  END,
  updated_at = now()
WHERE rate_limit_per_minute = 60;

CREATE TABLE IF NOT EXISTS channel_rate_limit_windows (
  channel TEXT NOT NULL,
  window_start TIMESTAMPTZ NOT NULL,
  request_count INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (channel, window_start)
);
