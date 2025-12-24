# 🎛️ Locust Web Dashboard Guide

## 🚀 Quick Start - Dashboard Mode

### Step 1: Start Dashboard

```bash
# Make script executable
chmod +x start_locust_dashboard.sh

# Start Locust Web UI
./start_locust_dashboard.sh
```

**ATAU manual:**
```bash
locust -f locustfile_auto.py
```

### Step 2: Open Browser

Buka browser dan akses:
```
http://localhost:8089
```

### Step 3: Configure di Dashboard

Di halaman dashboard, isi form:

1. **Number of users (peak concurrency)**
   ```
   Contoh: 1000
   ```
   Jumlah virtual users yang akan simulasikan traffic

2. **Spawn rate (users started/second)**
   ```
   Contoh: 50
   ```
   Berapa cepat users di-spawn (50 = spawn 50 users per detik)

3. **Host (URL to load test)**
   ```
   Contoh: http://localhost:8080
   atau: https://your-deployed-api.com
   ```
   URL server yang mau di-test

4. **Run time (Optional)**
   ```
   Contoh: 10m
   Format: 5s, 1m, 10m, 1h
   ```
   Kosongkan jika mau manual stop

### Step 4: Click "Start Swarming"

Load test akan langsung jalan!

---

## 📊 Dashboard Features

### Real-Time Metrics

Dashboard menampilkan:

#### 1. **Statistics Table**
- **Type**: Request method (GET, POST, dll)
- **Name**: Endpoint name
- **Requests**: Total requests
- **Fails**: Jumlah gagal
- **Median**: Response time median (ms)
- **90%ile**: 90th percentile response time
- **95%ile**: 95th percentile response time
- **99%ile**: 99th percentile response time
- **Average**: Average response time
- **Min/Max**: Response time range
- **Average size**: Response size
- **Current RPS**: Requests per second (real-time)

#### 2. **Charts Tab**
Grafik real-time:
- 📈 **Total Requests per Second** (RPS over time)
- 📉 **Response Times** (median, 95th percentile)
- 👥 **Number of Users** (active users over time)

#### 3. **Failures Tab**
Daftar errors:
- Error message
- Number of occurrences
- Affected endpoint

#### 4. **Exceptions Tab**
Stack traces untuk debugging

#### 5. **Current Ratio Tab**
Distribusi request per endpoint

---

## 🎮 Interactive Controls

### While Test is Running

**Stop Button**: Hentikan test kapan saja

**Edit Button**: Ubah jumlah users on-the-fly
- Bisa increase/decrease users tanpa stop test
- Useful untuk spike testing

**Reset Stats**: Clear current statistics

**Download Data**: Export results
- Download CSV
- Download stats history

---

## 💡 Common Scenarios

### Scenario 1: Quick Smoke Test

```
Number of users: 10
Spawn rate: 2
Host: http://localhost:8080
Run time: 1m
```

### Scenario 2: Medium Load Test

```
Number of users: 500
Spawn rate: 25
Host: http://localhost:8080
Run time: 5m
```

### Scenario 3: Heavy Stress Test

```
Number of users: 2000
Spawn rate: 100
Host: http://localhost:8080
Run time: 10m
```

### Scenario 4: Spike Test (Manual)

```
1. Start dengan:
   - Users: 100
   - Spawn rate: 10

2. Saat test running, click "Edit"

3. Increase to:
   - Users: 1000
   - Spawn rate: 200

4. Watch bagaimana system handle spike!
```

### Scenario 5: Remote Server Testing

```
Number of users: 1000
Spawn rate: 50
Host: https://your-deployed-api.com
Run time: 10m
```

---

## 📸 Dashboard Screenshot Guide

### Main Dashboard
```
┌─────────────────────────────────────────────────────┐
│  LOCUST                                             │
│                                                     │
│  State: RUNNING        Users: 1000/1000            │
│  Total RPS: 2,500      Failures: 0.05%             │
│                                                     │
│  [Statistics] [Charts] [Failures] [Exceptions]     │
│                                                     │
│  ┌───────────────────────────────────────────────┐ │
│  │ Type │ Name              │ # reqs │ Median   │ │
│  ├──────┼───────────────────┼────────┼──────────┤ │
│  │ POST │ /sessions/create  │ 10,000 │ 45ms     │ │
│  │ GET  │ /sessions/status  │ 50,000 │ 12ms     │ │
│  │ POST │ /langchain/exec.. │ 5,000  │ 230ms    │ │
│  └──────┴───────────────────┴────────┴──────────┘ │
│                                                     │
│  [STOP] [EDIT] [RESET] [DOWNLOAD]                 │
└─────────────────────────────────────────────────────┘
```

---

## 🎯 Step-by-Step Example

### Complete Dashboard Workflow

**1. Start Dashboard**
```bash
./start_locust_dashboard.sh
```

**2. Open Browser**
- Navigate to: `http://localhost:8089`
- Lihat halaman start form

**3. Fill Form**
```
Number of users: 1000
Spawn rate: 50
Host: http://localhost:8080
```

**4. Click "Start Swarming"**
- Test mulai jalan
- Users di-spawn gradually (50/second)
- Metrics update real-time

**5. Monitor Metrics**
- Tab "Statistics": Lihat response times
- Tab "Charts": Lihat grafik RPS
- Tab "Failures": Check errors (idealnya 0%)

**6. During Test**
- Watch RPS stabilize
- Monitor response times
- Check for errors/failures

**7. Click "Stop" When Done**
- Or wait for run time to complete

**8. Download Results**
- Click "Download Data"
- Save CSV untuk analisis

**9. View Report**
- Click "Download Report"
- HTML report dengan charts

---

## ⚙️ Advanced Settings

### Custom Locust Configuration

Edit `locustfile_auto.py` untuk customize:

```python
class AutomatedSessionUser(HttpUser):
    # Change wait time between requests
    wait_time = between(1, 3)  # Wait 1-3 seconds
    
    # Change task weights
    @task(10)  # Higher number = run more often
    def create_session(self):
        # ...
```

### Multiple Locustfiles

Bisa switch locustfile:

```bash
# Use automated flow (with test endpoint)
locust -f locustfile_auto.py

# Use partial flow (without test endpoint)
locust -f locustfile.py
```

---

## 🔧 Troubleshooting

### Dashboard tidak open

**Check:**
```bash
# Cek port 8089 free
lsof -i :8089

# Kill jika ada process lain
kill -9 <PID>
```

### Can't connect to API

**Check:**
```bash
# Test API running
curl http://localhost:8080/health

# Check firewall
# Check URL spelling
```

### High failure rate

**Common causes:**
- API overloaded (reduce users)
- Test endpoint not enabled (`APP_ENV=testing`)
- Database connection limit reached
- Network issues

---

## 📱 Dashboard on Mobile/Remote

### Access from Other Devices

```bash
# Start Locust with bind to all interfaces
locust -f locustfile_auto.py --host=http://localhost:8080 --web-host=0.0.0.0

# Access from other device on same network:
http://YOUR_IP:8089
# Example: http://192.168.1.100:8089
```

---

## 💾 Save and Share Results

### Export Options

**1. Download Statistics CSV**
- Click "Download Data" → "Download CSV"
- Contains all metrics in spreadsheet format

**2. Download HTML Report**
- Click "Download Data" → "Download report"
- Full HTML report with charts

**3. Screenshot Dashboard**
- Use browser screenshot
- Capture charts during test

**4. Save Configuration**
Create preset files:

```bash
# config/load-test-small.json
{
  "users": 100,
  "spawn_rate": 10,
  "host": "http://localhost:8080",
  "run_time": "2m"
}
```

---

## 🎓 Tips & Best Practices

### 1. Start Small, Scale Up
```
First run: 10 users
Second run: 100 users
Third run: 500 users
Final run: 1000+ users
```

### 2. Watch the Graphs
- RPS should be steady
- Response times should be consistent
- Failures should be < 1%

### 3. Use Edit Feature
- Start with low users
- Gradually increase during test
- Observe breaking point

### 4. Save Important Runs
- Download CSV for analysis
- Screenshot interesting patterns
- Document findings

### 5. Test Different Times
- Test during different times of day
- Compare results
- Identify patterns

---

## 🚀 Quick Reference Commands

```bash
# Start dashboard (automated flow)
./start_locust_dashboard.sh

# Or manual
locust -f locustfile_auto.py

# Start dashboard (partial flow)
locust -f locustfile.py

# Start with custom port
locust -f locustfile_auto.py --web-port=8090

# Start with default host
locust -f locustfile_auto.py --host=https://api.example.com

# Access dashboard
open http://localhost:8089  # Mac
xdg-open http://localhost:8089  # Linux
```

---

## 📊 Keyboard Shortcuts in Dashboard

- **W**: Start swarming
- **S**: Stop test
- **R**: Reset stats
- **ESC**: Close dialogs

---

Selamat testing dengan dashboard! 🎉
Jauh lebih mudah dan interaktif! 🎛️
