-- Demo seed for local PostgreSQL mode.

INSERT INTO conversations (id, customer, channel, intent, status, risk_level, last_message)
VALUES
  ('conv_demo_refund', '林夏', 'Web', '退款规则', 'Active', 'LOW', '7 天无理由退货的运费怎么计算？'),
  ('conv_demo_transfer', '周辰', 'WeChat', '人工协助', 'NeedsHuman', 'HIGH', '我已经催了三次，必须马上找人工处理。')
ON CONFLICT (id) DO UPDATE
SET
  customer = EXCLUDED.customer,
  channel = EXCLUDED.channel,
  intent = EXCLUDED.intent,
  status = EXCLUDED.status,
  risk_level = EXCLUDED.risk_level,
  last_message = EXCLUDED.last_message;

INSERT INTO knowledge_articles (id, title, category, content, tags, trust_level)
VALUES
  ('kb_refund_7d', '7 天无理由退货', '售后', '签收后 7 天内可申请无理由退货；非质量问题由用户承担寄回运费，质量问题由商家承担。', ARRAY['退款', '退货', '运费'], 'HIGH'),
  ('kb_invoice', '电子发票开具', '订单', '订单完成后可在订单详情申请电子发票，通常 24 小时内发送到预留邮箱。', ARRAY['发票', '订单'], 'HIGH'),
  ('kb_vip_transfer', '高风险投诉转人工', '服务规则', '出现强烈投诉、法律风险、连续催办等信号时，Agent 应停止自由生成并转人工。', ARRAY['投诉', '人工', '风险'], 'HIGH')
ON CONFLICT (id) DO UPDATE
SET
  title = EXCLUDED.title,
  category = EXCLUDED.category,
  content = EXCLUDED.content,
  tags = EXCLUDED.tags,
  trust_level = EXCLUDED.trust_level,
  updated_at = now();

INSERT INTO agent_rules (id, code, name, trigger_expr, action, enabled)
VALUES
  ('rule_low_evidence', 'NO_EVIDENCE_FALLBACK', '无可靠证据兜底', 'knowledge evidence empty', 'safe_fallback_and_create_gap', true),
  ('rule_human_transfer', 'TRANSFER_THRESHOLD', '转人工阈值', '投诉/催办/法律风险', 'recommend_human_transfer', true)
ON CONFLICT (code) DO UPDATE
SET
  name = EXCLUDED.name,
  trigger_expr = EXCLUDED.trigger_expr,
  action = EXCLUDED.action,
  enabled = EXCLUDED.enabled,
  updated_at = now();

INSERT INTO conversation_messages (id, conversation_id, role, content, engine, safe, fallback_reason, evidence_ids)
VALUES
  ('msg_demo_1', 'conv_demo_refund', 'user', '7 天无理由退货的运费怎么计算？', 'customer', true, '', ARRAY[]::TEXT[]),
  ('msg_demo_2', 'conv_demo_refund', 'assistant', '根据售后知识库，签收 7 天内可申请无理由退货；非质量问题寄回运费通常由用户承担，质量问题由商家承担。', 'rag+rule', true, '', ARRAY['kb_refund_7d'])
ON CONFLICT (id) DO UPDATE
SET
  content = EXCLUDED.content,
  engine = EXCLUDED.engine,
  safe = EXCLUDED.safe,
  fallback_reason = EXCLUDED.fallback_reason,
  evidence_ids = EXCLUDED.evidence_ids;
