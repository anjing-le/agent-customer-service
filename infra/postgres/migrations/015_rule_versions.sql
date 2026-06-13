ALTER TABLE agent_rules
  ADD COLUMN IF NOT EXISTS version TEXT NOT NULL DEFAULT '2026-06-active',
  ADD COLUMN IF NOT EXISTS stage TEXT NOT NULL DEFAULT 'active',
  ADD COLUMN IF NOT EXISTS hit_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_hit_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_agent_rules_stage ON agent_rules (stage, enabled, code);

UPDATE agent_rules
SET version = '2026-06-active',
    stage = 'active'
WHERE stage = '';

INSERT INTO agent_rules (id, code, name, trigger_expr, action, enabled, version, stage)
VALUES (
  'rule_cancel_canary',
  'CANCEL_RISK_TRANSFER',
  '取消/退订灰度转人工',
  '取消订单/退订服务/退款争议',
  'recommend_human_transfer',
  true,
  '2026-06-canary',
  'canary'
)
ON CONFLICT (code) DO UPDATE
SET
  name = EXCLUDED.name,
  trigger_expr = EXCLUDED.trigger_expr,
  action = EXCLUDED.action,
  enabled = EXCLUDED.enabled,
  version = EXCLUDED.version,
  stage = EXCLUDED.stage,
  updated_at = now();
