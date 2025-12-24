-- Seed script for load testing
-- Creates test users and pre-connected sessions

-- Insert test user
INSERT INTO users (user_id, api_key, created_at, updated_at)
VALUES ('test_user', 'test_key', NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING;

-- Insert pre-connected sessions for testing
-- These sessions simulate already-connected WhatsApp accounts
INSERT INTO sessions (
    user_id, 
    agent_id, 
    agent_name, 
    phone_number, 
    status, 
    langchain_url,
    connected_at, 
    created_at, 
    updated_at
)
VALUES 
    ('test_user', 'test_agent_1', 'Test Agent 1', '628123456789', 'connected', 
     'https://api.example.com', NOW(), NOW(), NOW()),
    ('test_user', 'test_agent_2', 'Test Agent 2', '628987654321', 'connected', 
     'https://api.example.com', NOW(), NOW(), NOW()),
    ('test_user', 'test_agent_3', 'Test Agent 3', '628111222333', 'connected', 
     'https://api.example.com', NOW(), NOW(), NOW())
ON CONFLICT (user_id, agent_id) DO UPDATE 
SET 
    status = 'connected', 
    connected_at = NOW(),
    updated_at = NOW();

-- Verify insertion
SELECT 
    agent_id, 
    agent_name, 
    status, 
    phone_number, 
    connected_at 
FROM sessions 
WHERE user_id = 'test_user';
