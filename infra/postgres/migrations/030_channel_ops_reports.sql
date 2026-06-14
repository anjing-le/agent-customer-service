CREATE TABLE IF NOT EXISTS channel_ops_reports (
  id TEXT PRIMARY KEY,
  format TEXT NOT NULL,
  content_type TEXT NOT NULL,
  content TEXT NOT NULL,
  summary JSONB NOT NULL DEFAULT '{}'::jsonb,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_channel_ops_reports_generated_at
  ON channel_ops_reports (generated_at DESC, id DESC);
