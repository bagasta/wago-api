# Load Testing Guide

## Setup Locust

### 1. Install Locust

```bash
pip install locust
```

### 2. Run Load Tests

#### Basic Test (All Endpoints)
```bash
locust -f locustfile.py --host=http://localhost:8080
```

Then open http://localhost:8089 in your browser to configure:
- Number of users (start with 10-50)
- Spawn rate (start with 5-10 users/second)

#### Focused Session Testing
```bash
locust -f locustfile.py --host=http://localhost:8080 FocusedSessionUser
```

#### Focused Langchain Testing
```bash
locust -f locustfile.py --host=http://localhost:8080 FocusedLangchainUser
```

#### Headless Mode (No Web UI)
```bash
locust -f locustfile.py --host=http://localhost:8080 \
  --users 100 \
  --spawn-rate 10 \
  --run-time 5m \
  --headless \
  --html report.html
```

## Strategy for QR Code Challenge

### Problem
The `/sessions/create` endpoint requires QR code scanning to fully connect, which can't be automated in load testing.

### Solutions Implemented

#### 1. **Test Partial Flow** (Default Strategy)
- Load test only tests QR generation, not full connection
- Measures: API response time, QR generation speed, DB write performance
- Status codes 200 (success) and 409 (conflict/already exists) are both considered success

#### 2. **Mock Endpoint for Testing** (Recommended Addition)
Add a test-only endpoint that bypasses QR scanning:

**File**: `internal/delivery/http/handler/session_handler.go`

```go
// CreateTestSession - FOR TESTING ONLY
// Bypass QR scanning and set session as connected
func (h *SessionHandler) CreateTestSession(c *fiber.Ctx) error {
    // Only enable in test/dev environment
    if os.Getenv("APP_ENV") != "development" && os.Getenv("APP_ENV") != "testing" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Endpoint only available in test/dev mode",
        })
    }
    
    // Implementation here...
}
```

Add to router:
```go
if config.IsTesting() {
    sessions.Post("/create-test", sessionHandler.CreateTestSession)
}
```

#### 3. **Pre-seed Connected Sessions**
For testing endpoints that require connected sessions:

**SQL Script**: `tests/seed_test_sessions.sql`

```sql
-- Insert test user
INSERT INTO users (user_id, api_key, created_at, updated_at)
VALUES ('test_user', 'test_key', NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING;

-- Insert pre-connected sessions
INSERT INTO sessions (
    user_id, agent_id, agent_name, 
    phone_number, status, 
    connected_at, created_at, updated_at
)
VALUES 
    ('test_user', 'test_agent_1', 'Test Agent 1', '628123456789', 'connected', NOW(), NOW(), NOW()),
    ('test_user', 'test_agent_2', 'Test Agent 2', '628987654321', 'connected', NOW(), NOW(), NOW())
ON CONFLICT (user_id, agent_id) DO UPDATE 
SET status = 'connected', connected_at = NOW();
```

Run before testing:
```bash
psql -h localhost -U postgres -d whatsapp_api -f tests/seed_test_sessions.sql
```

## Load Testing Scenarios

### Scenario 1: API Stress Test
**Goal**: Test maximum throughput

```bash
locust -f locustfile.py --host=http://localhost:8080 \
  --users 500 \
  --spawn-rate 50 \
  --run-time 10m
```

### Scenario 2: Sustained Load Test
**Goal**: Test stability over time

```bash
locust -f locustfile.py --host=http://localhost:8080 \
  --users 100 \
  --spawn-rate 10 \
  --run-time 1h
```

### Scenario 3: Spike Test
**Goal**: Test behavior during traffic spikes

```bash
# Start with baseline
locust -f locustfile.py --host=http://localhost:8080 \
  --users 50 \
  --spawn-rate 10

# Then manually increase to 500 users via web UI
# Observe behavior and recovery
```

## Metrics to Monitor

### Application Metrics
- Response time (p50, p90, p95, p99)
- Requests per second
- Error rate
- Success rate per endpoint

### System Metrics (Monitor separately)
```bash
# CPU and Memory
htop

# Database connections
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Application logs
tail -f logs/app.log
```

### Database Performance
```sql
-- Active queries
SELECT pid, usename, state, query 
FROM pg_stat_activity 
WHERE state != 'idle';

-- Table sizes
SELECT 
    tablename, 
    pg_size_pretty(pg_total_relation_size(tablename::text)) 
FROM pg_tables 
WHERE schemaname = 'public';
```

## Expected Results

### Target Performance (from PRD)
- API response time: < 200ms (p95)
- QR generation: < 500ms
- Support: 1000+ concurrent connections
- Throughput: 10,000+ requests/minute

### Interpreting Results

#### Good Performance
- Response time < 200ms for 95% of requests
- Error rate < 1%
- Stable memory usage
- No database connection exhaustion

#### Warning Signs
- Response time > 500ms
- Error rate > 5%
- Memory leaks (increasing over time)
- Database connection pool exhaustion

## Cleanup After Testing

```bash
# Remove test sessions via API
curl -X DELETE http://localhost:8080/api/v1/sessions/delete \
  -H "Content-Type: application/json" \
  -d '{"agentId": "test_agent_1"}'

# Or cleanup via SQL
psql -c "DELETE FROM sessions WHERE agent_id LIKE 'agent_test_%';"
```

## Troubleshooting

### High Error Rate
- Check application logs
- Verify database connection pool settings
- Check system resources (CPU, memory, disk)

### Slow Response Times
- Check database query performance
- Monitor database connection pool
- Check for N+1 query problems

### Connection Refused
- Verify API is running
- Check firewall settings
- Verify correct host:port in Locust config
