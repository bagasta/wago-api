# 🐳 Production Deployment Guide - Docker + Nginx + Load Balancing

## Overview

Panduan deployment production-ready untuk WhatsApp API dengan:
- **Docker** untuk containerization
- **Nginx** sebagai reverse proxy & load balancer
- **PostgreSQL** database
- Support untuk **ribuan concurrent users**
- Optimized untuk VPS IP: **194.238.23.242**

---

## 🏗️ Architecture

```
Internet
    ↓
194.238.23.242:80 (Nginx)
    ↓
[Load Balancer]
    ↓
┌────────┬────────┬────────┐
│ API #1 │ API #2 │ API #3 │  (Multiple API instances)
└────────┴────────┴────────┘
         ↓
    PostgreSQL
```

---

## 📋 Prerequisites

### 1. VPS Requirements

**Minimum untuk 1000+ users:**
- CPU: 4 cores
- RAM: 8GB
- Storage: 50GB SSD
- OS: Ubuntu 20.04+ atau Debian 11+

**Recommended untuk 5000+ users:**
- CPU: 8 cores
- RAM: 16GB
- Storage: 100GB SSD

### 2. Software Requirements

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose -y

# Install Nginx
sudo apt install nginx -y

# Verify installations
docker --version
docker-compose --version
nginx -v
```

---

## 🚀 Step-by-Step Deployment

### Step 1: Setup Project di VPS

```bash
# SSH ke VPS
ssh root@194.238.23.242

# Clone project
git clone <your-repo-url> /opt/whatsapp-api
cd /opt/whatsapp-api

# Atau upload via scp/rsync
rsync -avz --exclude 'venv' --exclude 'node_modules' \
  ~/Whatsapp-API-GO/ root@194.238.23.242:/opt/whatsapp-api/
```

### Step 2: Configure Environment

```bash
cd /opt/whatsapp-api

# Copy dan edit .env
cp .env.example .env
nano .env
```

**`.env` untuk Production:**
```bash
# Application
APP_ENV=production          # ← PRODUCTION mode (test endpoint disabled)
APP_PORT=8080
APP_NAME=WhatsApp-API

# Database (akan connect ke Docker PostgreSQL)
DB_HOST=postgres            # ← Container name
DB_PORT=5432
DB_USER=wago
DB_PASSWORD=your-secure-password-here  # ← GANTI!
DB_NAME=whatsapp_api
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=100      # ← Tinggi untuk ribuan users

# WhatsApp
WA_AUTO_RECONNECT=true
WA_QR_TIMEOUT=60
WA_LOG_LEVEL=INFO

# Langchain
LANGCHAIN_DEFAULT_TIMEOUT=30s
LANGCHAIN_BASE_URL=https://your-langchain-url.com

# Security
API_KEY_HEADER=Authorization
RATE_LIMIT_REQUESTS=10000   # ← Tinggi untuk load
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Step 3: Create docker-compose.yml untuk Production

```bash
nano docker-compose.production.yml
```

**`docker-compose.production.yml`:**
```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: whatsapp_postgres
    restart: always
    environment:
      POSTGRES_DB: whatsapp_api
      POSTGRES_USER: wago
      POSTGRES_PASSWORD: your-secure-password-here  # GANTI!
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U wago"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - whatsapp_network

  # API Instance 1
  api1:
    build: .
    container_name: whatsapp_api_1
    restart: always
    env_file:
      - .env
    environment:
      - DB_HOST=postgres
    ports:
      - "8081:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs
    networks:
      - whatsapp_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # API Instance 2 (untuk load balancing)
  api2:
    build: .
    container_name: whatsapp_api_2
    restart: always
    env_file:
      - .env
    environment:
      - DB_HOST=postgres
    ports:
      - "8082:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs
    networks:
      - whatsapp_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # API Instance 3 (untuk load balancing)
  api3:
    build: .
    container_name: whatsapp_api_3
    restart: always
    env_file:
      - .env
    environment:
      - DB_HOST=postgres
    ports:
      - "8083:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs
    networks:
      - whatsapp_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  whatsapp_network:
    driver: bridge

volumes:
  postgres_data:
```

### Step 4: Create Optimized Dockerfile

```bash
nano Dockerfile
```

**`Dockerfile`:**
```dockerfile
# Multi-stage build for smaller image
FROM golang:1.24-alpine AS builder

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o whatsapp-api cmd/api/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/whatsapp-api .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/config ./config

# Create logs directory
RUN mkdir -p /app/logs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run
CMD ["./whatsapp-api"]
```

### Step 5: Configure Nginx sebagai Load Balancer

```bash
sudo nano /etc/nginx/sites-available/whatsapp-api
```

**`/etc/nginx/sites-available/whatsapp-api`:**
```nginx
# Upstream servers (API instances)
upstream whatsapp_api {
    # Load balancing method: least_conn (best for websockets/long-polling)
    least_conn;
    
    # API instances
    server localhost:8081 max_fails=3 fail_timeout=30s;
    server localhost:8082 max_fails=3 fail_timeout=30s;
    server localhost:8083 max_fails=3 fail_timeout=30s;
    
    # Health check
    keepalive 32;
}

# Rate limiting zones
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_conn_zone $binary_remote_addr zone=addr:10m;

server {
    listen 80;
    listen [::]:80;
    
    # Server menggunakan IP VPS
    server_name 194.238.23.242;
    
    # Logging
    access_log /var/log/nginx/whatsapp-api-access.log;
    error_log /var/log/nginx/whatsapp-api-error.log;
    
    # Client body size (untuk upload files via WhatsApp)
    client_max_body_size 50M;
    
    # Timeouts
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
    
    # Rate limiting
    limit_req zone=api_limit burst=200 nodelay;
    limit_conn addr 100;
    
    # Proxy headers
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # Health check endpoint (tidak di-rate limit)
    location /health {
        proxy_pass http://whatsapp_api;
        access_log off;
    }
    
    # API endpoints
    location /api/ {
        proxy_pass http://whatsapp_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # Swagger documentation
    location /swagger/ {
        proxy_pass http://whatsapp_api;
    }
    
    # Root
    location / {
        proxy_pass http://whatsapp_api;
    }
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
```

**Enable site:**
```bash
# Create symlink
sudo ln -s /etc/nginx/sites-available/whatsapp-api /etc/nginx/sites-enabled/

# Remove default
sudo rm /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Step 6: Deploy!

```bash
cd /opt/whatsapp-api

# Build dan start containers
docker-compose -f docker-compose.production.yml up -d --build

# Check logs
docker-compose -f docker-compose.production.yml logs -f

# Check health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health

# Test via Nginx
curl http://194.238.23.242/health
```

---

## 🔒 Security Hardening

### 1. Firewall (UFW)

```bash
# Enable UFW
sudo ufw enable

# Allow SSH (IMPORTANT!)
sudo ufw allow 22/tcp

# Allow HTTP
sudo ufw allow 80/tcp

# Allow HTTPS (untuk future SSL)
sudo ufw allow 443/tcp

# Block direct access ke API ports
sudo ufw deny 8081/tcp
sudo ufw deny 8082/tcp
sudo ufw deny 8083/tcp

# Allow dari localhost
sudo ufw allow from 127.0.0.1 to any port 8081
sudo ufw allow from 127.0.0.1 to any port 8082
sudo ufw allow from 127.0.0.1 to any port 8083

# Check status
sudo ufw status
```

### 2. Fail2ban (Protection dari brute force)

```bash
# Install
sudo apt install fail2ban -y

# Configure
sudo nano /etc/fail2ban/jail.local
```

**`/etc/fail2ban/jail.local`:**
```ini
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/whatsapp-api-error.log

[nginx-limit-req]
enabled = true
port = http,https
logpath = /var/log/nginx/whatsapp-api-error.log
maxretry = 10
```

```bash
# Restart fail2ban
sudo systemctl restart fail2ban
sudo systemctl enable fail2ban
```

### 3. Automated Backups

```bash
# Create backup script
sudo nano /opt/backup-db.sh
```

**`/opt/backup-db.sh`:**
```bash
#!/bin/bash

BACKUP_DIR="/opt/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Backup database
docker exec whatsapp_postgres pg_dump -U wago whatsapp_api | gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: db_backup_$DATE.sql.gz"
```

```bash
# Make executable
sudo chmod +x /opt/backup-db.sh

# Setup cron (daily at 2 AM)
sudo crontab -e
```

Add:
```
0 2 * * * /opt/backup-db.sh >> /var/log/backup.log 2>&1
```

---

## 📊 Monitoring & Logging

### 1. Setup Log Rotation

```bash
sudo nano /etc/logrotate.d/whatsapp-api
```

```
/var/log/nginx/whatsapp-api-*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 www-data adm
    sharedscripts
    postrotate
        systemctl reload nginx > /dev/null 2>&1
    endscript
}

/opt/whatsapp-api/logs/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0644 root root
}
```

### 2. Monitor Resources

```bash
# Install htop
sudo apt install htop -y

# Monitor in real-time
htop

# Check Docker stats
docker stats

# Check Nginx status
sudo systemctl status nginx

# Check disk usage
df -h
```

### 3. Check Logs

```bash
# Nginx access log
sudo tail -f /var/log/nginx/whatsapp-api-access.log

# Nginx error log
sudo tail -f /var/log/nginx/whatsapp-api-error.log

# Docker logs
docker-compose -f docker-compose.production.yml logs -f api1
docker-compose -f docker-compose.production.yml logs -f postgres

# System logs
sudo journalctl -u nginx -f
```

---

## 🚀 Performance Tuning

### 1. Nginx Optimization

```bash
sudo nano /etc/nginx/nginx.conf
```

**Key settings:**
```nginx
user www-data;
worker_processes auto;  # Auto-detect CPU cores
pid /run/nginx.pid;

events {
    worker_connections 4096;  # Increase untuk ribuan connections
    use epoll;
    multi_accept on;
}

http {
    # Basic Settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    server_tokens off;  # Hide Nginx version
    
    # Buffer sizes
    client_body_buffer_size 128k;
    client_max_body_size 50m;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 8k;
    
    # Timeouts
    client_body_timeout 12;
    client_header_timeout 12;
    send_timeout 10;
    
    # Gzip Compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript 
               application/json application/javascript application/xml+rss;
    
    # Connection settings
    keepalive_requests 100;
    
    # Include other configs
    include /etc/nginx/mime.types;
    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
```

### 2. PostgreSQL Tuning

```bash
# Edit PostgreSQL config in Docker
docker exec -it whatsapp_postgres sh
```

Or create custom config and mount it:

**`postgres.conf`:**
```
# Connection Settings
max_connections = 200
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 10MB
min_wal_size = 1GB
max_wal_size = 4GB
```

### 3. Go API Optimization (Already in code)

Environment variables di `.env`:
```bash
DB_MAX_CONNECTIONS=100
DB_MAX_IDLE_CONNECTIONS=25
RATE_LIMIT_REQUESTS=10000
RATE_LIMIT_WINDOW=1m
```

---

## 🧪 Load Testing Production

### From Local Machine

```bash
# Test ke VPS
locust -f locustfile.py --host=http://194.238.23.242

# Or headless
locust -f locustfile.py \
  --host=http://194.238.23.242 \
  --users=5000 \
  --spawn-rate=100 \
  --run-time=10m \
  --headless \
  --html=production_load_test.html
```

### Monitoring During Test

**Terminal 1 (VPS):**
```bash
# Watch resources
htop
```

**Terminal 2 (VPS):**
```bash
# Watch Docker stats
watch -n 1 docker stats
```

**Terminal 3 (VPS):**
```bash
# Watch logs
sudo tail -f /var/log/nginx/whatsapp-api-access.log
```

**Terminal 4 (VPS):**
```bash
# Watch active connections
watch -n 1 'netstat -an | grep :80 | wc -l'
```

---

## 🔧 Troubleshooting

### Issue 1: High CPU Usage

```bash
# Check which container uses most CPU
docker stats

# Scale down if needed
docker-compose -f docker-compose.production.yml scale api1=1 api2=1 api3=1

# Or add more instances
docker-compose -f docker-compose.production.yml up -d --scale api1=2
```

### Issue 2: Database Connection Pool Exhausted

```bash
# Increase in .env
DB_MAX_CONNECTIONS=200

# Restart containers
docker-compose -f docker-compose.production.yml restart
```

### Issue 3: Nginx 502 Bad Gateway

```bash
# Check if backends are up
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health

# Check Docker logs
docker-compose -f docker-compose.production.yml logs api1

# Restart Nginx
sudo systemctl restart nginx
```

### Issue 4: Out of Memory

```bash
# Check memory
free -h

# Add swap (temporary)
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

---

## 📈 Scaling Further

### Horizontal Scaling (More API Instances)

Edit `docker-compose.production.yml`, add more services:

```yaml
  api4:
    # ... sama seperti api1, api2, api3
    ports:
      - "8084:8080"
```

Update Nginx upstream:
```nginx
upstream whatsapp_api {
    least_conn;
    server localhost:8081;
    server localhost:8082;
    server localhost:8083;
    server localhost:8084;  # New instance
}
```

### Vertical Scaling (Bigger VPS)

Upgrade VPS resources dan update PostgreSQL config accordingly.

---

## ✅ Production Checklist

Before going live:

- [ ] `.env` configured dengan production settings
- [ ] `APP_ENV=production` (test endpoint disabled)
- [ ] Database password diganti (strong password)
- [ ] Firewall enabled (UFW)
- [ ] Fail2ban configured
- [ ] Backup script setup (cron)
- [ ] Log rotation configured
- [ ] Nginx SSL certificate (untuk HTTPS) - Optional tapi recommended
- [ ] Load test completed successfully
- [ ] Monitoring setup
- [ ] Health checks working
- [ ] Documentation updated

---

## 🎯 Expected Performance

Dengan setup ini (3 API instances, 8GB RAM VPS):

- **Concurrent Users**: 3,000 - 5,000
- **Requests/Second**: 2,000 - 3,000
- **Response Time**: < 200ms (p95)
- **Error Rate**: < 0.5%
- **Uptime**: 99.9%

Untuk lebih dari 5,000 users: Add more API instances atau scale VPS.

---

## 📞 Maintenance Commands

```bash
# Restart all services
docker-compose -f docker-compose.production.yml restart

# Update code
git pull
docker-compose -f docker-compose.production.yml up -d --build

# View logs
docker-compose -f docker-compose.production.yml logs -f

# Backup database
/opt/backup-db.sh

# Restore database
gunzip < /opt/backups/db_backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i whatsapp_postgres psql -U wago whatsapp_api

# Clean old images
docker system prune -a

# Check disk space
df -h
du -sh /var/lib/docker
```

---

Deployment siap untuk production dengan scalability untuk ribuan users! 🚀
