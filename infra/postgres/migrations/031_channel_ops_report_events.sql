CREATE TABLE IF NOT EXISTS channel_ops_report_events (
  id TEXT PRIMARY KEY,
  action TEXT NOT NULL,
  actor TEXT NOT NULL,
  status TEXT NOT NULL,
  report_id TEXT NOT NULL DEFAULT '',
  format TEXT NOT NULL,
  pruned INTEGER NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_channel_ops_report_events_created_at
  ON channel_ops_report_events (created_at DESC, id DESC);
