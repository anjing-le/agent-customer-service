CREATE TABLE IF NOT EXISTS rule_release_events (
  id TEXT PRIMARY KEY,
  rule_code TEXT NOT NULL,
  version TEXT NOT NULL,
  action TEXT NOT NULL,
  actor TEXT NOT NULL DEFAULT 'operator',
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rule_release_events_rule ON rule_release_events (rule_code, created_at DESC);
