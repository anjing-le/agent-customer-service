ALTER TABLE transfer_tickets
  ADD COLUMN IF NOT EXISTS channel TEXT NOT NULL DEFAULT 'Web';

UPDATE transfer_tickets
SET channel = conversations.channel
FROM conversations
WHERE transfer_tickets.conversation_id = conversations.id
  AND transfer_tickets.channel = 'Web';

CREATE TABLE IF NOT EXISTS channel_policies (
  channel TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  tone TEXT NOT NULL,
  sla_minutes INTEGER NOT NULL,
  risk_boost TEXT NOT NULL DEFAULT 'NORMAL',
  escalation_note TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT true,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO channel_policies (channel, display_name, tone, sla_minutes, risk_boost, escalation_note, enabled)
VALUES
  ('Web', 'Web 客服', '标准、清晰、可追溯', 30, 'NORMAL', '网页渠道按标准客服 SLA 处理。', true),
  ('WeChat', '微信客服', '简洁、安抚、快速接管', 15, 'HIGH', '微信投诉和催办要更快进入人工队列。', true),
  ('App', 'App 客服', '直接、产品化、引导自助', 20, 'NORMAL', 'App 内问题优先引导订单和售后入口。', true),
  ('Marketplace', '平台店铺客服', '谨慎、合规、避免承诺', 10, 'HIGH', '平台投诉可能影响店铺指标，优先升级。', true)
ON CONFLICT (channel) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  tone = EXCLUDED.tone,
  sla_minutes = EXCLUDED.sla_minutes,
  risk_boost = EXCLUDED.risk_boost,
  escalation_note = EXCLUDED.escalation_note,
  enabled = EXCLUDED.enabled,
  updated_at = now();
