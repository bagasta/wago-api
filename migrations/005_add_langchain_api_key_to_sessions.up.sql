ALTER TABLE sessions
ADD COLUMN IF NOT EXISTS langchain_api_key TEXT;
