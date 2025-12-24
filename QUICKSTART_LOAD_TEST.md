# 🚀 Quick Start - Load Testing (Step by Step)

Panduan lengkap menjalankan load test dari NOL sampai selesai!

## 📋 Prerequisites

Pastikan sudah terinstall:
- ✅ Go 1.24+
- ✅ PostgreSQL
- ✅ Python 3.8+ (untuk Locust)
- ✅ pip (Python package manager)

## 🎯 Step-by-Step Guide

### STEP 1: Setup Database

```bash
# Jika belum ada database, buat dulu
psql -U postgres -c "CREATE DATABASE whatsapp_api;"

# Atau kalau pakai Docker
docker-compose up -d postgres
```

### STEP 2: Konfigurasi Environment

```bash
# Copy file .env.example ke .env
cp .env.example .env

# Edit .env file, tambahkan line ini:
nano .env  # atau gunakan text editor favorit
```

**Tambahkan di `.env`:**
```bash
# Application
APP_ENV=testing          # ← PENTING! Untuk aktifkan test endpoint
APP_PORT=8080
APP_NAME=WhatsApp-API

# Database
DB_HOST=localhost
DB_PORT=5432            # atau 5544 jika pakai docker
DB_USER=postgres        # atau wago jika pakai docker
DB_PASSWORD=your_pass   # atau wago-pass jika pakai docker
DB_NAME=whatsapp_api
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=50

# WhatsApp
WA_AUTO_RECONNECT=true
WA_QR_TIMEOUT=60
WA_LOG_LEVEL=INFO

# Langchain
LANGCHAIN_DEFAULT_TIMEOUT=30s
LANGCHAIN_BASE_URL=https://api.example.com

# Security
API_KEY_HEADER=Authorization
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

**Yang PALING PENTING:** Pastikan `APP_ENV=testing` agar test endpoint aktif!

### STEP 3: Install Dependencies Go

```bash
# Install Go dependencies
go mod download
go mod tidy
```

### STEP 4: Jalankan API Server

```bash
# Terminal 1: Jalankan API server
go run cmd/api/main.go
```

Tunggu sampai muncul:
```
Server starting on :8080
```

**Test apakah server running:**
```bash
# Di terminal baru
curl http://localhost:8080/health
```

Harus return: `{"status":"ok"}`

### STEP 5: Install Locust (Load Testing Tool)

```bash
# Terminal 2: Install Locust
pip install -r requirements-test.txt

# Atau install manual
pip install locust
```

**Verify installation:**
```bash
locust --version
```

### STEP 6: Jalankan Load Test! 🎉

Ada 3 cara:

#### **Cara 1: Script Otomatis (RECOMMENDED)**

```bash
# Jalankan script otomatis
./run_load_test.sh
```

Script akan:
- ✅ Check semua dependencies
- ✅ Check API server running
- ✅ Tanya mau seed test sessions atau tidak
- ✅ Run load test
- ✅ Generate report HTML
- ✅ Tanya mau cleanup atau tidak

**Custom jumlah users:**
```bash
# 100 users (smoke test)
USERS=100 SPAWN_RATE=10 RUN_TIME=2m ./run_load_test.sh

# 1000 users
USERS=1000 SPAWN_RATE=50 RUN_TIME=5m ./run_load_test.sh

# 5000 users (stress test!)
USERS=5000 SPAWN_RATE=100 RUN_TIME=10m ./run_load_test.sh
```

#### **Cara 2: Headless Mode (Tanpa Web UI)**

```bash
locust -f locustfile_auto.py \
  --host=http://localhost:8080 \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=5m \
  --headless \
  --html=report.html
```

#### **Cara 3: Web UI Mode (Manual Control)**

```bash
# Start Locust web UI
locust -f locustfile_auto.py --host=http://localhost:8080
```

Then:
1. Buka browser: http://localhost:8089
2. Set number of users: `1000`
3. Set spawn rate: `50`
4. Click **Start swarming**
5. Watch real-time metrics!

### STEP 7: Monitor Test

**Saat test berjalan, monitor di terminal:**

```bash
# Terminal 3: Watch logs
tail -f logs/app.log  # jika ada

# Atau watch database
watch -n 2 'psql -U postgres -d whatsapp_api -c "SELECT COUNT(*) FROM sessions;"'
```

**Metrics yang akan muncul:**
- Requests per second (RPS)
- Response time (median, p95, p99)
- Error rate
- Active users

### STEP 8: Lihat Results

**HTML Report:**
```bash
# Jika pakai script otomatis
open load_test_reports/load_test_*.html  # Mac
xdg-open load_test_reports/load_test_*.html  # Linux

# Jika pakai cara 2
open report.html
```

**Report berisi:**
- 📊 Response time statistics
- 📈 Request per second chart
- 🎯 Success/failure rate
- 📉 Response time distribution
- 📝 Detailed per-endpoint metrics

### STEP 9: Cleanup (Optional)

```bash
# Hapus test sessions dari database
psql -U postgres -d whatsapp_api -c "DELETE FROM sessions WHERE agent_id LIKE 'load_test_%';"

# Atau via API
curl -X DELETE http://localhost:8080/api/v1/sessions/delete \
  -H "Content-Type: application/json" \
  -d '{"agentId": "load_test_xxx"}'
```

## 🎬 Complete Example (Copy-Paste Ready!)

```bash
# 1. Setup environment
cp .env.example .env
echo "APP_ENV=testing" >> .env

# 2. Install dependencies
go mod download
pip install -r requirements-test.txt

# 3. Start API server (Terminal 1)
go run cmd/api/main.go &

# Wait 3 seconds
sleep 3

# 4. Verify server
curl http://localhost:8080/health

# 5. Run load test (100 users quick test)
USERS=100 SPAWN_RATE=10 RUN_TIME=2m ./run_load_test.sh

# 6. View results
xdg-open load_test_reports/load_test_*.html
```

## 🔍 Troubleshooting

### ❌ Error: "Endpoint only available in test/development mode"

**Solusi:**
```bash
# Pastikan APP_ENV=testing di .env
grep APP_ENV .env

# Jika tidak ada, tambahkan
echo "APP_ENV=testing" >> .env

# Restart API server
pkill -f "go run"
go run cmd/api/main.go
```

### ❌ Error: "Connection refused"

**Solusi:**
```bash
# Check apakah API running
curl http://localhost:8080/health

# Jika tidak, start API
go run cmd/api/main.go
```

### ❌ Error: "locust: command not found"

**Solusi:**
```bash
# Install Locust
pip install locust

# Atau
pip install -r requirements-test.txt
```

### ❌ Error: Database connection failed

**Solusi:**
```bash
# Check PostgreSQL running
psql -U postgres -c "SELECT version();"

# Atau start via docker
docker-compose up -d postgres

# Check connection settings di .env
cat .env | grep DB_
```

### ❌ High error rate during test

**Kemungkinan penyebab:**
1. Database connection pool habis → Tingkatkan `DB_MAX_CONNECTIONS` di .env
2. CPU/Memory penuh → Kurangi jumlah users
3. API timeout → Tingkatkan timeout settings

## 📊 Expected Results (Target Performance)

Dari PRD, target performance:
- ✅ Response time: < 200ms (p95)
- ✅ QR generation: < 500ms
- ✅ Support: 1000+ concurrent users
- ✅ Throughput: 10,000+ requests/minute
- ✅ Error rate: < 1%

## 🎯 Test Scenarios

### Scenario 1: Smoke Test (Cepat)
```bash
USERS=50 SPAWN_RATE=10 RUN_TIME=1m ./run_load_test.sh
```

### Scenario 2: Load Test (Normal)
```bash
USERS=1000 SPAWN_RATE=50 RUN_TIME=5m ./run_load_test.sh
```

### Scenario 3: Stress Test (Heavy)
```bash
USERS=5000 SPAWN_RATE=100 RUN_TIME=10m ./run_load_test.sh
```

### Scenario 4: Endurance Test (Long)
```bash
USERS=500 SPAWN_RATE=25 RUN_TIME=1h ./run_load_test.sh
```

## 💡 Pro Tips

1. **Start Small**: Mulai dengan 100 users, naikkan bertahap
2. **Monitor Resources**: Watch CPU, Memory selama test
3. **Check Logs**: Lihat error patterns di logs
4. **Cleanup**: Hapus test sessions setelah test
5. **Iterate**: Run multiple tests dengan setting berbeda

## 📞 Need Help?

Jika ada error atau pertanyaan, check:
- `tests/AUTOMATED_LOAD_TESTING.md` - Dokumentasi lengkap
- `tests/load_testing_guide.md` - Manual testing guide
- Logs di terminal

Happy Load Testing! 🚀
