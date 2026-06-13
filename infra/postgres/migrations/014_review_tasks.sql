CREATE TABLE IF NOT EXISTS review_tasks (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL REFERENCES conversation_messages(id),
  conversation_id TEXT NOT NULL REFERENCES conversations(id),
  channel TEXT NOT NULL DEFAULT 'Web',
  assignee TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'OPEN',
  priority TEXT NOT NULL DEFAULT 'NORMAL',
  reason TEXT NOT NULL DEFAULT 'Agent 回复抽检',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_review_tasks_status ON review_tasks (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_review_tasks_message ON review_tasks (message_id);

INSERT INTO review_tasks (id, message_id, conversation_id, channel, status, priority, reason)
VALUES ('review_demo_2', 'msg_demo_2', 'conv_demo_refund', 'Web', 'OPEN', 'HIGH', '种子回复需验证证据完整性')
ON CONFLICT (id) DO NOTHING;
