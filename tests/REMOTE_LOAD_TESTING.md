# 🌐 Load Testing untuk Deployed Server

Panduan load testing untuk server yang sudah di-deploy (production/staging).

## 🎯 Quick Start - Remote Server

### Method 1: Script Otomatis dengan URL Remote

```bash
# Set URL deployment kamu
export API_HOST=https://your-api.example.com

# Jalankan load test
USERS=1000 SPAWN_RATE=50 RUN_TIME=5m ./run_load_test.sh
```

### Method 2: Direct Locust Command

```bash
# Ganti dengan URL deployment kamu
locust -f locustfile_auto.py \
  --host=https://your-api.example.com \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=5m \
  --headless \
  --html=report.html
```

### Method 3: Web UI (Recommended untuk Testing Bertahap)

```bash
# Start Locust Web UI
locust -f locustfile_auto.py

# Buka browser: http://localhost:8089
# Masukkan:
#   - Host: https://your-api.example.com
#   - Number of users: 1000
#   - Spawn rate: 50
```

## ⚠️ PENTING: Test Endpoint di Production

### Scenario 1: Production dengan Test Endpoint Enabled

Jika server production punya `APP_ENV=testing` atau `development`:

✅ **Bisa pakai `locustfile_auto.py`** (full automation)
- Gunakan endpoint `/sessions/create-test`
- Session otomatis "connected" tanpa QR

```bash
# Test langsung
locust -f locustfile_auto.py \
  --host=https://your-api.example.com \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=5m \
  --headless
```

### Scenario 2: Production Tanpa Test Endpoint

Jika server production `APP_ENV=production` (test endpoint disabled):

⚠️ **Pakai `locustfile.py`** (partial flow)
- Test QR generation endpoint
- Tidak test full connection (karena perlu scan manual)

```bash
# Test partial flow
locust -f locustfile.py \
  --host=https://your-api.example.com \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=5m \
  --headless
```

**ATAU** Pre-seed sessions:

```bash
# Connect ke database production (HATI-HATI!)
psql -h production-db.example.com -U postgres -d whatsapp_api \
  -f tests/seed_test_sessions.sql

# Kemudian test dengan sessions yang sudah ada
locust -f locustfile_auto.py --host=https://your-api.example.com
```

## 🌍 Load Testing dari Local ke Remote

### Setup

```bash
# 1. Pastikan Locust terinstall
pip install -r requirements-test.txt

# 2. Test connectivity
curl https://your-api.example.com/health

# 3. Jalankan test
API_HOST=https://your-api.example.com \
USERS=100 \
SPAWN_RATE=10 \
RUN_TIME=2m \
./run_load_test.sh
```

### Example URLs

```bash
# Production
API_HOST=https://api.yourproject.com ./run_load_test.sh

# Staging
API_HOST=https://staging-api.yourproject.com ./run_load_test.sh

# Development
API_HOST=https://dev-api.yourproject.com ./run_load_test.sh

# Local Docker
API_HOST=http://localhost:8080 ./run_load_test.sh
```

## 📊 Recommended Test Scenarios untuk Deployed Server

### 1. Smoke Test (Quick Health Check)
```bash
# Cepat, minimal load
locust -f locustfile_auto.py \
  --host=https://your-api.example.com \
  --users=10 \
  --spawn-rate=2 \
  --run-time=1m \
  --headless
```

### 2. Load Test (Normal Traffic)
```bash
# Simulasi traffic normal
API_HOST=https://your-api.example.com \
USERS=500 \
SPAWN_RATE=25 \
RUN_TIME=5m \
./run_load_test.sh
```

### 3. Stress Test (Peak Traffic)
```bash
# Simulasi peak hours
API_HOST=https://your-api.example.com \
USERS=2000 \
SPAWN_RATE=100 \
RUN_TIME=10m \
./run_load_test.sh
```

### 4. Spike Test (Traffic Surge)
```bash
# Start web UI
locust -f locustfile_auto.py --host=https://your-api.example.com

# Di web UI:
# 1. Start dengan 100 users
# 2. Wait 2 minutes
# 3. Increase ke 1000 users (spike!)
# 4. Monitor behavior
```

### 5. Endurance Test (Long Duration)
```bash
# Test stability over time
API_HOST=https://your-api.example.com \
USERS=500 \
SPAWN_RATE=25 \
RUN_TIME=1h \
./run_load_test.sh
```

## 🔒 Security Considerations

### 1. API Keys & Authentication
```python
# Edit locustfile_auto.py jika perlu auth khusus
class AutomatedSessionUser(HttpUser):
    def on_start(self):
        self.api_key = "your-production-api-key"  # Bisa dari env var
        # atau
        import os
        self.api_key = os.getenv("API_KEY", "default-key")
```

### 2. Rate Limiting
Jika production punya rate limiting:
```bash
# Spawn lebih pelan
SPAWN_RATE=10  # Slower spawn rate

# Atau increase wait time
# Edit locustfile_auto.py:
wait_time = between(2, 5)  # Slower request rate
```

### 3. Test Endpoint Security
**BEST PRACTICE**: Disable test endpoint di production!

```bash
# Di production server, set:
APP_ENV=production

# Maka endpoint /create-test akan return 403
```

## 📈 Monitoring During Remote Testing

### 1. Monitor Server Metrics
```bash
# SSH ke server
ssh user@your-server.com

# Monitor resources
htop
docker stats  # jika pakai docker

# Monitor logs
tail -f /var/log/whatsapp-api/app.log
journalctl -u whatsapp-api -f  # jika pakai systemd
```

### 2. Database Monitoring
```bash
# Monitor active connections
psql -h db.example.com -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Monitor slow queries
psql -h db.example.com -U postgres -c \
  "SELECT query, state FROM pg_stat_activity WHERE state != 'idle';"
```

### 3. Application Metrics
Jika punya monitoring tools (Prometheus, Grafana, DataDog, dll):
- Watch CPU usage
- Monitor memory consumption
- Track response times
- Check error rates
- Monitor database connections

## 🎯 Performance Targets

Dari PRD, target untuk production:
- ✅ Response time: < 200ms (p95)
- ✅ Throughput: > 10,000 requests/minute
- ✅ Error rate: < 1%
- ✅ Availability: 99.9% uptime
- ✅ Concurrent users: 1000+

## 🚨 Warning Signs

Stop test jika melihat:
- ❌ Error rate > 10%
- ❌ Response time > 5 seconds
- ❌ Server CPU > 90% sustained
- ❌ Database connection pool exhausted
- ❌ Memory leaks (increasing over time)
- ❌ Disk space critical

## 📊 Analyzing Results

### Good Performance Indicators
```
✓ Response time: p50 < 100ms, p95 < 200ms, p99 < 500ms
✓ Error rate: < 1%
✓ RPS: Consistent throughout test
✓ Resource usage: Stable (no leaks)
```

### Warning Indicators
```
⚠ Response time: p95 > 500ms
⚠ Error rate: 1-5%
⚠ RPS: Declining over time
⚠ Memory: Slowly increasing
```

### Critical Issues
```
❌ Response time: p95 > 2s
❌ Error rate: > 5%
❌ RPS: Dropping significantly
❌ Server: Out of memory/CPU maxed
```

## 💡 Best Practices

### 1. Start Small
```bash
# Begin with smoke test
USERS=10 RUN_TIME=1m ./run_load_test.sh

# Gradually increase
USERS=100 RUN_TIME=2m ./run_load_test.sh
USERS=500 RUN_TIME=5m ./run_load_test.sh
USERS=1000 RUN_TIME=10m ./run_load_test.sh
```

### 2. Test from Different Locations
```bash
# Run from multiple regions if possible
# Region 1: US East
locust --host=https://api.example.com

# Region 2: Europe
locust --host=https://api.example.com

# Region 3: Asia
locust --host=https://api.example.com
```

### 3. Coordinate with Team
- 📅 Schedule tests during off-peak hours
- 📢 Notify team before load testing
- 🔔 Monitor alerts during test
- 📊 Share results with team

### 4. Cleanup After Testing
```bash
# Remove test sessions
curl -X DELETE https://your-api.example.com/api/v1/sessions/delete \
  -H "Content-Type: application/json" \
  -d '{"agentId": "load_test_xxx"}'

# Or via database
psql -h db.example.com -U postgres -d whatsapp_api -c \
  "DELETE FROM sessions WHERE agent_id LIKE 'load_test_%';"
```

## 📋 Pre-Test Checklist

- [ ] Notifikasi team akan ada load test
- [ ] Check deployment status (healthy)
- [ ] Backup database (optional tapi recommended)
- [ ] Set up monitoring dashboards
- [ ] Prepare runbook untuk rollback
- [ ] Test connectivity: `curl https://your-api.example.com/health`
- [ ] Verify API key/auth working
- [ ] Schedule during off-peak hours
- [ ] Have team on standby

## 📝 Post-Test Checklist

- [ ] Cleanup test sessions
- [ ] Archive HTML reports
- [ ] Document findings
- [ ] Share results dengan team
- [ ] Create tickets untuk issues found
- [ ] Update capacity planning docs
- [ ] Note any config changes needed

## Example: Complete Remote Load Test

```bash
#!/bin/bash
# Complete load test workflow for remote server

# 1. Set deployment URL
export API_HOST=https://api.yourproject.com

# 2. Pre-test health check
echo "Checking server health..."
curl $API_HOST/health || exit 1

# 3. Smoke test (quick validation)
echo "Running smoke test..."
USERS=10 SPAWN_RATE=2 RUN_TIME=1m ./run_load_test.sh

# 4. Load test (normal traffic)
echo "Running load test..."
USERS=500 SPAWN_RATE=25 RUN_TIME=5m ./run_load_test.sh

# 5. Stress test (peak traffic)
echo "Running stress test..."
USERS=2000 SPAWN_RATE=100 RUN_TIME=10m ./run_load_test.sh

# 6. Cleanup
echo "Cleaning up test sessions..."
psql -h db.example.com -U postgres -d whatsapp_api -c \
  "DELETE FROM sessions WHERE agent_id LIKE 'load_test_%';"

echo "Load testing completed! Check reports in load_test_reports/"
```

Selamat load testing server production! 🚀
