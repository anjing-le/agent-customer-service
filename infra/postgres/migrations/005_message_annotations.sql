CREATE TABLE IF NOT EXISTS message_annotations (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL REFERENCES conversation_messages(id),
  reviewer TEXT NOT NULL DEFAULT 'operator',
  verdict TEXT NOT NULL DEFAULT 'PASS',
  note TEXT NOT NULL DEFAULT '',
  groundedness INTEGER NOT NULL DEFAULT 5,
  safety INTEGER NOT NULL DEFAULT 5,
  helpfulness INTEGER NOT NULL DEFAULT 5,
  tags TEXT[] NOT NULL DEFAULT '{}',
  score INTEGER NOT NULL DEFAULT 100,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_message_annotations_message ON message_annotations (message_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_message_annotations_verdict ON message_annotations (verdict, created_at DESC);
