# 🔧 Troubleshooting Guide - Locust Test 403 Error

## Problem: `/create-test` endpoint returns 403 "Test endpoint not enabled"

### Root Cause
The endpoint checks `os.Getenv("APP_ENV")` at runtime. If it's not "testing" or "development", it returns 403.

---

## ✅ STEP-BY-STEP FIX

### 1. Verify `.env` File

```bash
cd /home/bagas/Whatsapp-API-GO

# Check if APP_ENV exists
grep "^APP_ENV" .env

# Should show EXACTLY:
# APP_ENV=testing
```

**If missing or wrong:**
```bash
# Fix it
echo "APP_ENV=testing" >> .env

# Or edit manually
nano .env
# Add line: APP_ENV=testing
```

### 2. Check if API Server is Running

```bash
# Check if process exists
ps aux | grep "go run cmd/api/main.go"

# Or try health check
curl http://localhost:8080/health
```

**If "Connection refused" or no process:**
```bash
# Start API server in NEW terminal
go run cmd/api/main.go
```

### 3. RESTART API Server

⚠️ **IMPORTANT:** `.env` file is only read when app STARTS!

```bash
# If API already running, STOP it first (Ctrl+C in that terminal)
# Then start again:
go run cmd/api/main.go
```

**Wait for:**
```
Server starting on :8080
```

### 4. Test Endpoint Manually

```bash
# Test the problematic endpoint
curl -X POST http://localhost:8080/api/v1/sessions/create-test \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "manual_test",
    "agentName": "Manual Test",
    "apiKey": "secret",
    "langchainUrl": "https://api.example.com"
  }'
```

**Expected SUCCESS response:**
```json
{
  "success": true,
  "message": "Test session created successfully (mock connected)",
  "data": {
    "sessionId": 1,
    "agentId": "manual_test",
    "status": "connected",
    "phoneNumber": "62811122233344",
    "connectedAt": "2024-12-24T14:12:00Z"
  }
}
```

**If still 403:**
```json
{
  "success": false,
  "error": "Endpoint only available in test/development mode"
}
```

→ Environment variable not loaded. Check step 5.

### 5. Debug Environment Loading

Create test file to verify:

```bash
# Create debug script
cat > debug_env.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "github.com/subosito/gotenv"
)

func main() {
    // Load .env
    gotenv.Load()
    
    // Check APP_ENV
    appEnv := os.Getenv("APP_ENV")
    fmt.Printf("APP_ENV = '%s'\n", appEnv)
    
    if appEnv == "testing" || appEnv == "development" {
        fmt.Println("✓ Test endpoint WILL BE enabled")
    } else {
        fmt.Println("✗ Test endpoint WILL BE disabled")
    }
}
EOF

# Run it
go run debug_env.go
```

**Should show:**
```
APP_ENV = 'testing'
✓ Test endpoint WILL BE enabled
```

### 6. Restart Locust Test

After API server is running with correct environment:

1. Stop current Locust test (Ctrl+C or STOP button)
2. Start fresh:
   ```bash
   source venv/bin/activate
   locust -f locustfile_auto.py
   ```
3. Open http://localhost:8089
4. Configure and start

---

## 📋 CHECKLIST

Before running load test:

- [ ] `.env` has `APP_ENV=testing`
- [ ] API server is running (`ps aux | grep "go run"`)
- [ ] API responds to `/health` (`curl http://localhost:8080/health`)
- [ ] `/create-test` endpoint works (`curl -X POST ...`)
- [ ] Locust can connect to API

---

## 🎯 QUICK FIX (Copy-Paste)

```bash
# Terminal 1: Fix and start API
cd /home/bagas/Whatsapp-API-GO

# Ensure APP_ENV is set
grep -q "^APP_ENV=testing" .env || echo "APP_ENV=testing" >> .env

# Start API
go run cmd/api/main.go
```

```bash
# Terminal 2: Start Locust (after API is up)
cd /home/bagas/Whatsapp-API-GO
source venv/bin/activate
locust -f locustfile_auto.py
```

---

## 🔍 If STILL Not Working

### Check Code in Handler

The endpoint code checks:

```go
// File: internal/delivery/http/handler/session_handler.go
func (h *SessionHandler) CreateTestSession(c *fiber.Ctx) error {
    env := os.Getenv("APP_ENV")
    if env != "development" && env != "testing" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "error":   "Endpoint only available in test/development mode",
        })
    }
    // ...
}
```

Make sure:
1. Code is saved (check file modification time)
2. No compilation errors
3. Server actually restarted (check server logs)

---

## 📊 Alternative: Use SQL Seed Instead

If test endpoint still problematic, use SQL approach:

```bash
# Pre-seed sessions in database
psql -h localhost -U postgres -d whatsapp_api -f tests/seed_test_sessions.sql

# Then use regular locustfile
locust -f locustfile.py
```

This bypasses the test endpoint completely!

---

## 💡 Pro Tip: Auto-Restart Script

Create `start_all.sh`:

```bash
#!/bin/bash

# Ensure APP_ENV
export APP_ENV=testing

# Start API in background
go run cmd/api/main.go &
API_PID=$!

# Wait for API to be ready
echo "Waiting for API..."
until curl -s http://localhost:8080/health > /dev/null; do
    sleep 1
done
echo "✓ API ready"

# Start Locust
source venv/bin/activate
locust -f locustfile_auto.py

# Cleanup on exit
kill $API_PID
```

Then just:
```bash
chmod +x start_all.sh
./start_all.sh
```

Done! 🎉
