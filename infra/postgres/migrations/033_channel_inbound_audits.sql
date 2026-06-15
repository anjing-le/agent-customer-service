CREATE TABLE IF NOT EXISTS channel_inbound_audits (
  id TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  external_conversation_id TEXT NOT NULL DEFAULT '',
  external_message_id TEXT NOT NULL DEFAULT '',
  origin TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  code TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  replay_key TEXT NOT NULL DEFAULT '',
  signature_preview TEXT NOT NULL DEFAULT '',
  content_hash TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_channel_inbound_audits_created_at
  ON channel_inbound_audits (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_channel_inbound_audits_channel_status
  ON channel_inbound_audits (channel, status, created_at DESC);
