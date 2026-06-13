CREATE TABLE IF NOT EXISTS transfer_tickets (
  id TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL,
  question TEXT NOT NULL,
  reason TEXT NOT NULL,
  priority TEXT NOT NULL DEFAULT 'HIGH',
  status TEXT NOT NULL DEFAULT 'OPEN',
  assignee TEXT NOT NULL DEFAULT '',
  resolution_note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_transfer_tickets_status ON transfer_tickets (status, priority, created_at DESC);

INSERT INTO transfer_tickets (id, conversation_id, question, reason, priority, status)
VALUES
  ('ticket_demo_transfer', 'conv_demo_transfer', '我已经催了三次，必须马上找人工处理。', 'TRANSFER_THRESHOLD', 'HIGH', 'OPEN')
ON CONFLICT (id) DO UPDATE
SET
  question = EXCLUDED.question,
  reason = EXCLUDED.reason,
  priority = EXCLUDED.priority,
  status = EXCLUDED.status;
