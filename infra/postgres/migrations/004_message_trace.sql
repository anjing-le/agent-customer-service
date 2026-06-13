ALTER TABLE conversation_messages
ADD COLUMN IF NOT EXISTS trace JSONB NOT NULL DEFAULT '{}'::jsonb;
