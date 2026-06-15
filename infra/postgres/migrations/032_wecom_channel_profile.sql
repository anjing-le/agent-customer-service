INSERT INTO channel_policies (channel, display_name, tone, sla_minutes, risk_boost, escalation_note, enabled)
VALUES
  ('WeCom', '企业微信客服', '专业、内部协同、保留上下文', 12, 'HIGH', '企业微信通常连接私域客户和内部协同，投诉和合同问题优先接管。', true)
ON CONFLICT (channel) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  tone = EXCLUDED.tone,
  sla_minutes = EXCLUDED.sla_minutes,
  risk_boost = EXCLUDED.risk_boost,
  escalation_note = EXCLUDED.escalation_note,
  enabled = EXCLUDED.enabled,
  updated_at = now();

INSERT INTO channel_integrations (
  channel,
  display_name,
  enabled,
  secret_source,
  secret_ref,
  signature_window_seconds,
  replay_protection,
  rotation_hint
)
VALUES
  ('WeCom', '企业微信客服', true, 'env', 'ANJING_CHANNEL_WECOM_SECRET', 300, true, '企业微信回调按 corpId + msgId 做对账并轮换 token')
ON CONFLICT (channel) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  enabled = EXCLUDED.enabled,
  secret_source = EXCLUDED.secret_source,
  secret_ref = EXCLUDED.secret_ref,
  signature_window_seconds = EXCLUDED.signature_window_seconds,
  replay_protection = EXCLUDED.replay_protection,
  rotation_hint = EXCLUDED.rotation_hint,
  updated_at = now();

UPDATE channel_integrations
SET next_secret_ref = 'ANJING_CHANNEL_WECOM_NEXT_SECRET'
WHERE channel = 'WeCom'
  AND EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'channel_integrations'
      AND column_name = 'next_secret_ref'
  );

UPDATE channel_integrations
SET allowed_origins = ARRAY['https://qyapi.weixin.qq.com']
WHERE channel = 'WeCom'
  AND EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'channel_integrations'
      AND column_name = 'allowed_origins'
  );

UPDATE channel_integrations
SET rate_limit_per_minute = 50
WHERE channel = 'WeCom'
  AND EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'channel_integrations'
      AND column_name = 'rate_limit_per_minute'
  );

INSERT INTO channel_alert_policies (channel, severity, threshold, window_minutes, notify_target, enabled)
VALUES
  ('WeCom', 'HIGH', 4, 60, 'wecom-oncall', true)
ON CONFLICT (channel) DO UPDATE
SET
  severity = EXCLUDED.severity,
  threshold = EXCLUDED.threshold,
  window_minutes = EXCLUDED.window_minutes,
  notify_target = EXCLUDED.notify_target,
  enabled = EXCLUDED.enabled,
  updated_at = now();
