CREATE TABLE IF NOT EXISTS channel_alert_policies (
  channel TEXT PRIMARY KEY,
  severity TEXT NOT NULL DEFAULT 'MEDIUM',
  threshold INTEGER NOT NULL DEFAULT 5,
  window_minutes INTEGER NOT NULL DEFAULT 60,
  notify_target TEXT NOT NULL DEFAULT 'ops-webhook',
  enabled BOOLEAN NOT NULL DEFAULT true,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO channel_alert_policies (channel, severity, threshold, window_minutes, notify_target, enabled)
VALUES
  ('Web', 'MEDIUM', 8, 60, 'ops-webhook', true),
  ('WeChat', 'HIGH', 5, 60, 'wechat-oncall', true),
  ('App', 'MEDIUM', 6, 60, 'app-oncall', true),
  ('Marketplace', 'HIGH', 3, 60, 'marketplace-oncall', true)
ON CONFLICT (channel) DO UPDATE
SET
  severity = EXCLUDED.severity,
  threshold = EXCLUDED.threshold,
  window_minutes = EXCLUDED.window_minutes,
  notify_target = EXCLUDED.notify_target,
  enabled = EXCLUDED.enabled,
  updated_at = now();
