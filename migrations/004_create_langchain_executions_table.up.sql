CREATE TABLE IF NOT EXISTS langchain_executions (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    user_message TEXT,
    langchain_response JSONB,
    execution_time_ms INTEGER,
    status VARCHAR(50),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_langchain_session ON langchain_executions(session_id);
CREATE INDEX IF NOT EXISTS idx_langchain_created ON langchain_executions(created_at DESC);
