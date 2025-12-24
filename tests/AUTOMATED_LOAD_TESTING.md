# Automated Load Testing Setup

## Quick Start - Ribuan Users Otomatis! 🚀

### 1. Set Environment untuk Testing
```bash
# Edit .env file
APP_ENV=testing  # atau development
```

### 2. Jalankan Script Otomatis
```bash
./run_load_test.sh
```

Script ini akan otomatis:
- ✓ Check dependencies (install Locust jika belum ada)
- ✓ Check API server
- ✓ Seed test sessions (optional)
- ✓ Run load test dengan ribuan users
- ✓ Generate HTML report
- ✓ Cleanup sessions (optional)

## Custom Configuration

### Test dengan Ribuan Users
```bash
# 1000 users
USERS=1000 SPAWN_RATE=50 RUN_TIME=10m ./run_load_test.sh

# 5000 users!
USERS=5000 SPAWN_RATE=100 RUN_TIME=15m ./run_load_test.sh

# Extreme: 10000 users
USERS=10000 SPAWN_RATE=200 RUN_TIME=20m ./run_load_test.sh
```

### Environment Variables
```bash
API_HOST=http://localhost:8080  # Target API
USERS=1000                      # Jumlah concurrent users
SPAWN_RATE=50                   # Users/second spawn rate
RUN_TIME=5m                     # Test duration (5m, 10m, 1h)
```

## Cara Kerja Otomatis

### Endpoint `/sessions/create-test`
Endpoint khusus untuk testing yang **bypass QR scanning**:

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/sessions/create-test \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "test_123",
    "agentName": "Test Agent",
    "apiKey": "secret",
    "langchainUrl": "https://api.example.com"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Test session created successfully (mock connected)",
  "data": {
    "sessionId": 1,
    "agentId": "test_123",
    "status": "connected",  // ← Langsung connected!
    "phoneNumber": "62811122233344",
    "connectedAt": "2024-12-24T11:00:00Z"
  }
}
```

### Keamanan
- ⚠️ Endpoint ini **HANYA aktif** jika `APP_ENV=testing` atau `APP_ENV=development`
- ⚠️ Di production (`APP_ENV=production`), endpoint ini return **403 Forbidden**
- ✓ Aman untuk deployment

## Manual Run (Tanpa Script)

### 1. Install Dependencies
```bash
pip install -r requirements-test.txt
```

### 2. Set Environment
```bash
# Edit .env
APP_ENV=testing
```

### 3. Run Locust
```bash
# Headless mode (otomatis)
locust -f locustfile_auto.py \
  --host=http://localhost:8080 \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=10m \
  --headless \
  --html=report.html

# Web UI mode (manual control)
locust -f locustfile_auto.py --host=http://localhost:8080
# Then open http://localhost:8089
```

## Test Scenarios

### Scenario 1: Quick Smoke Test
```bash
USERS=100 SPAWN_RATE=10 RUN_TIME=2m ./run_load_test.sh
```

### Scenario 2: Medium Load (Ribuan Users)
```bash
USERS=2000 SPAWN_RATE=100 RUN_TIME=10m ./run_load_test.sh
```

### Scenario 3: Stress Test (5000+ Users)
```bash
USERS=5000 SPAWN_RATE=200 RUN_TIME=15m ./run_load_test.sh
```

### Scenario 4: Endurance Test (Long Duration)
```bash
USERS=1000 SPAWN_RATE=50 RUN_TIME=1h ./run_load_test.sh
```

## Monitoring During Test

### 1. Application Metrics
```bash
# Terminal 1: Watch logs
tail -f logs/app.log

# Terminal 2: Monitor resources
htop
```

### 2. Database Performance
```bash
# Monitor active connections
watch -n 1 'psql -c "SELECT count(*) FROM pg_stat_activity WHERE state != '\''idle'\'';"'

# Monitor table sizes
psql -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::text)) FROM pg_tables WHERE schemaname = 'public';"
```

### 3. Locust Web UI (if not headless)
Open http://localhost:8089 to see real-time:
- Requests per second (RPS)
- Response times (median, 95th percentile)
- Failure rate
- Charts and graphs

## Reading Results

### HTML Report
File: `load_test_reports/load_test_YYYYMMDD_HHMMSS.html`

Contains:
- Request statistics per endpoint
- Response time distribution
- Charts and graphs
- Failure analysis

### CSV Files
Files: `load_test_reports/load_test_*_stats.csv`

For further analysis in Excel/Google Sheets:
- Per-endpoint metrics
- Time-series data
- Statistical aggregations

### Expected Performance (from PRD)
- ✅ Response time < 200ms (p95)
- ✅ Error rate < 1%
- ✅ Support 1000+ concurrent users
- ✅ Throughput > 10,000 req/min

## Cleanup

### Automatic Cleanup
Script will prompt you to cleanup test sessions after test.

### Manual Cleanup
```bash
# Via SQL
psql -h localhost -U postgres -d whatsapp_api -c \
  "DELETE FROM sessions WHERE agent_id LIKE 'load_test_%';"

# Via API (for each session)
curl -X DELETE http://localhost:8080/api/v1/sessions/delete \
  -H "Content-Type: application/json" \
  -d '{"agentId": "load_test_xxx"}'
```

## Troubleshooting

### Error: "Endpoint only available in test/development mode"
**Solution:** Set `APP_ENV=testing` in `.env` file

### Error: "Connection refused"
**Solution:** Make sure API server is running:
```bash
go run cmd/api/main.go
```

### Error: "Locust not found"
**Solution:** Install Locust:
```bash
pip install -r requirements-test.txt
```

### High error rate during test
**Possible causes:**
- Database connection pool exhausted → Increase `DB_MAX_CONNECTIONS` in `.env`
- CPU/Memory maxed out → Reduce concurrent users or increase resources
- Langchain API timeout → Increase `LANGCHAIN_DEFAULT_TIMEOUT`

### Slow response times
**Check:**
1. Database query performance
2. Connection pool settings
3. System resources (CPU, RAM, Disk I/O)
4. Network latency to external APIs

## Advanced: Docker Load Testing

### Run API in Docker
```bash
docker-compose up -d
```

### Run Locust in Docker
```bash
docker run --network=host -v $(pwd):/app -w /app locustio/locust \
  -f locustfile_auto.py \
  --host=http://localhost:8080 \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=10m \
  --headless
```

## Tips for Testing Ribuan Users

1. **Gradual Ramp-up**: Start dengan spawn rate rendah, naikkan bertahap
2. **Monitor Resources**: Watch CPU, Memory, Database connections
3. **Set Realistic Timeouts**: Jangan terlalu aggressive
4. **Cleanup Between Tests**: Delete test sessions untuk hasil konsisten
5. **Use Different Test Files**: 
   - `locustfile.py` → Partial flow (QR generation only)
   - `locustfile_auto.py` → Full automated flow

## Example: Testing 10,000 Users

```bash
# 1. Increase database connections
echo "DB_MAX_CONNECTIONS=100" >> .env

# 2. Set testing mode
echo "APP_ENV=testing" >> .env

# 3. Restart API
pkill -f "go run"
go run cmd/api/main.go &

# 4. Run massive load test
USERS=10000 SPAWN_RATE=200 RUN_TIME=30m ./run_load_test.sh

# 5. Analyze results
open load_test_reports/load_test_*.html
```

Selamat load testing! 🚀
