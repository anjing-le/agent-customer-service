CREATE TABLE IF NOT EXISTS channel_integrations (
  channel TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT true,
  secret_source TEXT NOT NULL DEFAULT 'env',
  secret_ref TEXT NOT NULL,
  signature_window_seconds INTEGER NOT NULL DEFAULT 300,
  replay_protection BOOLEAN NOT NULL DEFAULT true,
  rotation_hint TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

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
  ('Web', 'Web 客服', true, 'env', 'ANJING_CHANNEL_WEB_SECRET', 300, true, '按演示环境手动轮换 env secret'),
  ('WeChat', '微信客服', true, 'env', 'ANJING_CHANNEL_WECHAT_SECRET', 300, true, '生产接入时对齐微信回调 message id'),
  ('App', 'App 客服', true, 'env', 'ANJING_CHANNEL_APP_SECRET', 300, true, 'App 版本发布时同步轮换 secret'),
  ('Marketplace', '平台店铺客服', true, 'env', 'ANJING_CHANNEL_MARKETPLACE_SECRET', 300, true, '按平台回调密钥周期轮换')
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
