CREATE TABLE IF NOT EXISTS rule_approvals (
  id TEXT PRIMARY KEY,
  rule_code TEXT NOT NULL,
  approver TEXT NOT NULL DEFAULT 'qa-lead',
  risk_level TEXT NOT NULL DEFAULT 'LOW',
  sample_count INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'PENDING',
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rule_approvals_rule ON rule_approvals (rule_code, status, created_at DESC);
