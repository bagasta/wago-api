CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    agent_name VARCHAR(255),
    phone_number VARCHAR(50),
    qr_code TEXT,
    qr_code_base64 TEXT,
    session_data JSONB,
    status VARCHAR(50) DEFAULT 'disconnected',
    langchain_url TEXT,
    last_qr_generated_at TIMESTAMP,
    connected_at TIMESTAMP,
    disconnected_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, agent_id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_agent ON sessions(user_id, agent_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
