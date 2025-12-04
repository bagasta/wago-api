CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(255),
    from_number VARCHAR(50),
    to_number VARCHAR(50),
    message_text TEXT,
    message_type VARCHAR(50),
    direction VARCHAR(20),
    status VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_agent ON messages(agent_id);
CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at DESC);
