-- Agent Customer Service V1 schema.
-- IDs are text so teaching examples stay readable.

CREATE TABLE IF NOT EXISTS conversations (
  id TEXT PRIMARY KEY,
  customer TEXT NOT NULL,
  channel TEXT NOT NULL,
  intent TEXT NOT NULL DEFAULT '待识别',
  status TEXT NOT NULL DEFAULT 'Active',
  risk_level TEXT NOT NULL DEFAULT 'LOW',
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_message TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS conversation_messages (
  id TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL REFERENCES conversations(id),
  role TEXT NOT NULL,
  content TEXT NOT NULL,
  engine TEXT NOT NULL,
  safe BOOLEAN NOT NULL DEFAULT true,
  fallback_reason TEXT NOT NULL DEFAULT '',
  evidence_ids TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_articles (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  category TEXT NOT NULL,
  content TEXT NOT NULL,
  tags TEXT[] NOT NULL DEFAULT '{}',
  trust_level TEXT NOT NULL DEFAULT 'MEDIUM',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_gaps (
  id TEXT PRIMARY KEY,
  conversation_id TEXT NOT NULL,
  question TEXT NOT NULL,
  reason TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'OPEN',
  priority TEXT NOT NULL DEFAULT 'MEDIUM',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS agent_rules (
  id TEXT PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  trigger_expr TEXT NOT NULL,
  action TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT true,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations (status);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON conversation_messages (conversation_id, created_at);
CREATE INDEX IF NOT EXISTS idx_knowledge_gaps_status ON knowledge_gaps (status, priority);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_category ON knowledge_articles (category);
