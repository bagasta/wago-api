# 🐍 Tutorial Load Testing dengan Locust (Python Virtual Environment)

## 📋 Kenapa Pakai Virtual Environment?

Sistem Python modern (Ubuntu 23.04+, Debian 12+) menggunakan "externally-managed-environment" yang tidak mengizinkan install package langsung ke system Python. Solusinya: gunakan **virtual environment (venv)**.

---

## 🚀 Quick Start (Copy-Paste Ready!)

### Step 1: Buat dan Aktifkan Virtual Environment

```bash
# Navigate ke project directory
cd /home/bagas/Whatsapp-API-GO

# Buat virtual environment bernama 'venv'
python3 -m venv venv

# Aktifkan virtual environment
source venv/bin/activate

# Sekarang terminal prompt akan berubah jadi:
# (venv) user@hostname:~/Whatsapp-API-GO$
```

### Step 2: Install Locust di Virtual Environment

```bash
# Pastikan venv aktif (ada "(venv)" di prompt)
# Install Locust
pip install locust

# Atau install semua requirements
pip install -r requirements-test.txt
```

### Step 3: Jalankan Locust Dashboard

```bash
# Start Locust Web UI
locust -f locustfile_auto.py
```

### Step 4: Buka Browser

```
http://localhost:8089
```

**Done! 🎉** Dashboard Locust akan terbuka.

---

## 📖 Detailed Step-by-Step Guide

### Part 1: Setup Virtual Environment (Sekali Saja)

#### 1. Install python3-venv (jika belum ada)

```bash
# Check apakah python3-venv sudah ada
python3 -m venv --help

# Jika error, install:
sudo apt update
sudo apt install python3-venv python3-full
```

#### 2. Buat Virtual Environment

```bash
# Di root project
cd /home/bagas/Whatsapp-API-GO

# Buat venv
python3 -m venv venv

# Folder 'venv' akan dibuat berisi Python environment isolated
```

#### 3. Aktifkan Virtual Environment

```bash
# Aktivasi (Linux/Mac)
source venv/bin/activate

# Setelah aktif, prompt berubah:
# (venv) user@hostname:~/Whatsapp-API-GO$
```

**Catatan:** Setiap kali buka terminal baru, harus run `source venv/bin/activate` lagi!

#### 4. Upgrade pip (Recommended)

```bash
# Pastikan venv aktif
pip install --upgrade pip
```

#### 5. Install Locust dan Dependencies

```bash
# Install dari requirements file
pip install -r requirements-test.txt

# Atau install manual
pip install locust

# Verify installation
locust --version
# Output: locust 2.20.0 (atau versi lain)
```

### Part 2: Menjalankan Load Test

#### Option 1: Web Dashboard (Recommended)

```bash
# 1. Pastikan venv aktif
source venv/bin/activate

# 2. Start Locust
locust -f locustfile_auto.py

# 3. Buka browser
# http://localhost:8089

# 4. Input settings di web UI:
#    - Number of users: 1000
#    - Spawn rate: 50
#    - Host: http://localhost:8080 (atau URL deployed server)

# 5. Click "Start Swarming"
```

#### Option 2: Headless Mode (No Web UI)

```bash
# Pastikan venv aktif
source venv/bin/activate

# Run headless
locust -f locustfile_auto.py \
  --host=http://localhost:8080 \
  --users=1000 \
  --spawn-rate=50 \
  --run-time=5m \
  --headless \
  --html=report.html
```

#### Option 3: Menggunakan Script Otomatis

```bash
# Pastikan venv aktif
source venv/bin/activate

# Run script
API_HOST=http://localhost:8080 \
USERS=1000 \
SPAWN_RATE=50 \
RUN_TIME=5m \
./run_load_test.sh
```

### Part 3: Deaktivasi Virtual Environment

```bash
# Ketika selesai testing
deactivate

# Prompt kembali normal:
# user@hostname:~/Whatsapp-API-GO$
```

---

## 🔄 Workflow Sehari-hari

### Setiap Kali Mau Load Test:

```bash
# 1. Masuk ke project directory
cd /home/bagas/Whatsapp-API-GO

# 2. Aktifkan venv
source venv/bin/activate

# 3. Start Locust
locust -f locustfile_auto.py

# 4. Buka browser: http://localhost:8089

# 5. Setelah selesai, deactivate
deactivate
```

---

## 🛠️ Troubleshooting

### Problem 1: `python3 -m venv` tidak ditemukan

**Error:**
```
No module named venv
```

**Solution:**
```bash
sudo apt update
sudo apt install python3-venv python3-full
```

### Problem 2: `pip: command not found` di dalam venv

**Error:**
```
bash: pip: command not found
```

**Solution:**
```bash
# Deactivate dan recreate venv
deactivate
rm -rf venv
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip
```

### Problem 3: Lupa aktifkan venv

**Symptoms:**
- Command `locust` not found
- Atau error externally-managed-environment

**Solution:**
```bash
# Aktifkan venv dulu
source venv/bin/activate

# Lihat prompt berubah menjadi (venv)
```

### Problem 4: Virtual environment corrupt

**Solution:**
```bash
# Hapus dan buat ulang
rm -rf venv
python3 -m venv venv
source venv/bin/activate
pip install -r requirements-test.txt
```

### Problem 5: Permission denied saat buat venv

**Solution:**
```bash
# Pastikan di directory project kamu punya write access
ls -la ~/Whatsapp-API-GO/

# Jika perlu, fix permissions
chmod +w ~/Whatsapp-API-GO/
```

---

## 📁 Structure Setelah Setup

```
Whatsapp-API-GO/
├── venv/                    # ← Virtual environment (jangan commit!)
│   ├── bin/
│   │   ├── activate         # ← Script untuk activate
│   │   ├── locust           # ← Locust binary
│   │   └── python           # ← Python isolated
│   ├── lib/
│   └── pyvenv.cfg
├── locustfile_auto.py
├── requirements-test.txt
└── ...
```

**Catatan:** Folder `venv/` sudah ada di `.gitignore`, jadi tidak akan di-commit ke Git.

---

## 🎯 Complete Example Workflow

### First Time Setup:

```bash
# 1. Navigate ke project
cd ~/Whatsapp-API-GO

# 2. Buat venv
python3 -m venv venv

# 3. Aktifkan
source venv/bin/activate

# 4. Install dependencies
pip install -r requirements-test.txt

# 5. Verify
locust --version
```

### Daily Usage:

```bash
# Setiap kali mau test:

# 1. Aktifkan venv
cd ~/Whatsapp-API-GO
source venv/bin/activate

# 2. Start dashboard
locust -f locustfile_auto.py

# 3. Open http://localhost:8089

# 4. Setelah selesai
deactivate
```

---

## 💡 Pro Tips

### Tip 1: Buat Alias untuk Speed

Tambahkan ke `~/.bashrc` atau `~/.zshrc`:

```bash
# Alias untuk activate venv
alias activate-wago='cd ~/Whatsapp-API-GO && source venv/bin/activate'

# Alias untuk start locust
alias locust-start='cd ~/Whatsapp-API-GO && source venv/bin/activate && locust -f locustfile_auto.py'
```

Reload shell:
```bash
source ~/.bashrc
```

Sekarang cukup ketik:
```bash
locust-start
```

### Tip 2: Check Venv Active atau Tidak

```bash
# Jika venv aktif, akan ada "(venv)" di prompt:
(venv) user@host:~/path$

# Atau check dengan:
which python
# Output jika aktif: /home/user/Whatsapp-API-GO/venv/bin/python
# Output jika tidak: /usr/bin/python3
```

### Tip 3: Install Package Tambahan

```bash
# Selalu aktifkan venv dulu
source venv/bin/activate

# Baru install
pip install <package-name>

# Update requirements.txt
pip freeze > requirements-test.txt
```

### Tip 4: Multiple Terminal Windows

Jika pakai multiple terminal:

**Terminal 1** (API Server):
```bash
cd ~/Whatsapp-API-GO
go run cmd/api/main.go
```

**Terminal 2** (Locust):
```bash
cd ~/Whatsapp-API-GO
source venv/bin/activate
locust -f locustfile_auto.py
```

**Browser**:
```
http://localhost:8089
```

---

## 🔐 Security Notes

### Virtual Environment Best Practices

1. **Never commit venv/** to Git
   - Already in `.gitignore`
   - Recreate on each machine

2. **Use requirements.txt**
   - Commit requirements file
   - Others can recreate same environment

3. **Keep venv updated**
   ```bash
   source venv/bin/activate
   pip install --upgrade pip
   pip install --upgrade -r requirements-test.txt
   ```

---

## 📊 Testing Scenarios dengan venv

### Scenario 1: Local Testing

```bash
# Terminal 1: Start API
go run cmd/api/main.go

# Terminal 2: Start Locust
source venv/bin/activate
locust -f locustfile_auto.py

# Browser: http://localhost:8089
# Host: http://localhost:8080
# Users: 1000
# Spawn rate: 50
```

### Scenario 2: Remote Server Testing

```bash
# Aktifkan venv
source venv/bin/activate

# Start Locust
locust -f locustfile_auto.py

# Browser: http://localhost:8089
# Host: https://your-deployed-api.com
# Users: 1000
# Spawn rate: 50
```

### Scenario 3: Headless CI/CD

```bash
source venv/bin/activate

locust -f locustfile_auto.py \
  --host=https://staging-api.example.com \
  --users=500 \
  --spawn-rate=25 \
  --run-time=5m \
  --headless \
  --html=reports/load_test_$(date +%Y%m%d_%H%M%S).html
```

---

## 🎓 Additional Resources

### Learn More about venv:

```bash
# Python venv documentation
python3 -m venv --help

# List installed packages
source venv/bin/activate
pip list

# Show package info
pip show locust

# Check outdated packages
pip list --outdated
```

### Learn More about Locust:

```bash
# Locust help
source venv/bin/activate
locust --help

# Locust web interface options
locust -f locustfile_auto.py --web-port=8090

# Run in master-worker mode (advanced)
# Terminal 1 (master):
locust -f locustfile_auto.py --master

# Terminal 2+ (workers):
locust -f locustfile_auto.py --worker
```

---

## ✅ Quick Checklist

Sebelum mulai testing, pastikan:

- [ ] Virtual environment sudah dibuat (`venv/` folder exists)
- [ ] Virtual environment aktif (prompt shows `(venv)`)
- [ ] Locust terinstall di venv (`locust --version` works)
- [ ] API server running (`curl http://localhost:8080/health`)
- [ ] Browser ready untuk dashboard

---

## 🆘 Need Help?

Jika masih ada error:

1. **Check Python version**
   ```bash
   python3 --version
   # Should be 3.8+
   ```

2. **Recreate venv from scratch**
   ```bash
   rm -rf venv
   python3 -m venv venv
   source venv/bin/activate
   pip install -r requirements-test.txt
   ```

3. **Check system packages**
   ```bash
   sudo apt update
   sudo apt install python3-venv python3-pip python3-full
   ```

4. **Check file di docs/antigravity.md** untuk dokumentasi project

---

## 🎉 Ready to Go!

Sekarang kamu sudah siap untuk load testing dengan Locust menggunakan virtual environment!

**Quick Command Summary:**
```bash
# Setup (sekali)
python3 -m venv venv
source venv/bin/activate
pip install -r requirements-test.txt

# Daily usage (setiap kali test)
source venv/bin/activate
locust -f locustfile_auto.py
# Open: http://localhost:8089

# Done
deactivate
```

Happy Load Testing! 🚀
