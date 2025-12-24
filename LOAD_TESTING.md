# WhatsApp API - Load Testing Setup Otomatis

Setup lengkap untuk automated load testing dengan ribuan users!

## 📁 Files

### Testing Files
- **`locustfile_auto.py`** - Automated Locust yang bypass QR scanning
- **`run_load_test.sh`** - Script otomatis untuk run test
- **`tests/AUTOMATED_LOAD_TESTING.md`** - Panduan lengkap (BACA INI!)
- **`tests/seed_test_sessions.sql`** - SQL seed untuk pre-seeded sessions

### API Changes
- **New Endpoint**: `POST /api/v1/sessions/create-test`
  - Bypass QR scanning untuk testing
  - Hanya aktif jika `APP_ENV=testing` atau `development`
  - Return session yang langsung "connected"

## 🚀 Quick Start

### 1. Set Environment
```bash
echo "APP_ENV=testing" >> .env
```

### 2. Run Test
```bash
./run_load_test.sh
```

### 3. Custom Users
```bash
# 1000 users
USERS=1000 ./run_load_test.sh

# 5000 users
USERS=5000 SPAWN_RATE=100 ./run_load_test.sh

# 10000 users!
USERS=10000 SPAWN_RATE=200 RUN_TIME=20m ./run_load_test.sh
```

## 📊 What Gets Tested

✓ Session creation (via test endpoint)
✓ Session status checks
✓ Session details
✓ Langchain execution
✓ Health checks
✓ Automatic cleanup

## 📈 Reports

HTML reports saved to: `load_test_reports/load_test_TIMESTAMP.html`

## 🔒 Security

⚠️ Test endpoint **ONLY works** when `APP_ENV=testing` or `development`
✅ Production safe - returns 403 Forbidden

## 📖 Full Documentation

See: `tests/AUTOMATED_LOAD_TESTING.md`
