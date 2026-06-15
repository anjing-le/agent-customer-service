ALTER TABLE channel_alert_policies
  ADD COLUMN IF NOT EXISTS inbound_audit_min_samples INTEGER NOT NULL DEFAULT 3,
  ADD COLUMN IF NOT EXISTS inbound_audit_min_acceptance_rate INTEGER NOT NULL DEFAULT 80,
  ADD COLUMN IF NOT EXISTS inbound_audit_max_error_count INTEGER NOT NULL DEFAULT 2;

UPDATE channel_alert_policies
SET
  inbound_audit_min_samples = CASE
    WHEN channel IN ('Web', 'App') THEN 5
    WHEN channel IN ('WeChat', 'WeCom', 'Marketplace', 'Douyin', 'Xiaohongshu') THEN 4
    ELSE 3
  END,
  inbound_audit_min_acceptance_rate = CASE
    WHEN channel IN ('WeChat', 'WeCom', 'Marketplace', 'Douyin', 'Xiaohongshu') THEN 85
    ELSE 80
  END,
  inbound_audit_max_error_count = CASE
    WHEN inbound_audit_max_error_count <= 0 THEN 2
    ELSE inbound_audit_max_error_count
  END,
  updated_at = now();
