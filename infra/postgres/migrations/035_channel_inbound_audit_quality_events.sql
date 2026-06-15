CREATE TABLE IF NOT EXISTS channel_inbound_audit_quality_events (
  id TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  severity TEXT NOT NULL DEFAULT 'MEDIUM',
  status TEXT NOT NULL DEFAULT 'OPEN',
  failure_code TEXT NOT NULL,
  total INTEGER NOT NULL DEFAULT 0,
  accepted INTEGER NOT NULL DEFAULT 0,
  rejected INTEGER NOT NULL DEFAULT 0,
  acceptance_rate INTEGER NOT NULL DEFAULT 0,
  min_samples INTEGER NOT NULL DEFAULT 3,
  min_acceptance_rate INTEGER NOT NULL DEFAULT 80,
  max_error_count INTEGER NOT NULL DEFAULT 2,
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_channel_inbound_audit_quality_events_created_at
  ON channel_inbound_audit_quality_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_channel_inbound_audit_quality_events_channel
  ON channel_inbound_audit_quality_events (channel);
