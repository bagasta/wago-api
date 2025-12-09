# Panduan Deployment (Systemd + Nginx)

Panduan ini menjelaskan cara men-deploy aplikasi WhatsApp API Go ke server Linux (Ubuntu/Debian) menggunakan **Systemd** sebagai process manager dan **Nginx** sebagai reverse proxy.

## 1. Persiapan Server

Pastikan server sudah terinstall:
- Go (jika build di server) atau upload binary hasil build.
- PostgreSQL.
- Nginx.

```bash
# Update & Install Nginx
sudo apt update
sudo apt install nginx postgresql
```

## 2. Build Aplikasi

Jika Anda mem-build di lokal komputer, jalankan:
```bash
# Build untuk Linux (jika lokal Anda Windows/Mac)
GOOS=linux GOARCH=amd64 go build -o whatsapp-api cmd/api/main.go
```

Jika build langsung di server:
```bash
go build -o whatsapp-api cmd/api/main.go
```

## 3. Setup Direktori Aplikasi

Kita akan menaruh aplikasi di `/var/www/wago-api`.

```bash
# Buat folder
sudo mkdir -p /var/www/wago-api

# Copy binary dan config
sudo cp whatsapp-api /var/www/wago-api/
sudo cp -r config /var/www/wago-api/
sudo cp -r migrations /var/www/wago-api/
sudo cp .env /var/www/wago-api/

# Set permission (sesuaikan user, misal: root atau ubuntu)
sudo chown -R root:root /var/www/wago-api
sudo chmod +x /var/www/wago-api/whatsapp-api
```

Pastikan `.env` minimal berisi:
```
SERVER_PORT=9300
DATABASE_URL=postgresql://postgres:aiagronomists@194.238.23.242:5432/whatsapp_api?sslmode=disable
```
Port 9300 sengaja dipakai agar tidak bentrok dengan service lain, dan aplikasi otomatis mencari port kosong berikutnya jika 9300 sedang dipakai.

## 4. Konfigurasi Systemd

Systemd akan menjaga aplikasi tetap berjalan di background dan otomatis restart jika error.

1. Copy file service:
   ```bash
   sudo cp deploy/whatsapp-api.service /etc/systemd/system/wago-api.service
```

2. Edit file jika perlu (sesuaikan User, WorkingDirectory, dll):
```bash
sudo nano /etc/systemd/system/wago-api.service
```

3. Reload & Start Service:
```bash
sudo systemctl daemon-reload
sudo systemctl start wago-api
sudo systemctl enable wago-api # Agar auto-start saat boot
```

4. Cek Status:
```bash
sudo systemctl status wago-api
```

## 5. Konfigurasi Nginx (Reverse Proxy)

Nginx akan meneruskan request HTTPS (port 443) ke port 9300 (aplikasi). File `deploy/nginx.conf` sudah menyiapkan redirect HTTP -> HTTPS dan secara default menunjuk ke sertifikat Let’s Encrypt di `/etc/letsencrypt/live/wago-api.chiefaiofficer.id/` (ubah bila path berbeda).

1. Buat file config baru di Nginx:
```bash
sudo nano /etc/nginx/sites-available/wago-api
```
*(Copy isi dari file `deploy/nginx.conf` ke sini. Pastikan `server_name` sudah sesuai dan path sertifikat ada.)*

2. Aktifkan config:
```bash
sudo ln -s /etc/nginx/sites-available/wago-api /etc/nginx/sites-enabled/
```

3. Test & Restart Nginx:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 6. Selesai!

Sekarang API Anda sudah bisa diakses melalui domain atau IP server tanpa port 8080.
Contoh: `http://api.example.com/api/v1/sessions/create`

---

### Sertifikat SSL (Let's Encrypt)

Jika belum memiliki sertifikat, jalankan (atau sesuaikan sesuai DNS Anda):
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d wago-api.chiefaiofficer.id
```
Perintah tersebut akan membuatkan file `fullchain.pem` dan `privkey.pem` di `/etc/letsencrypt/live/wago-api.chiefaiofficer.id/`. Setelah tersedia, jangan lupa memastikan `ssl_certificate` dan `ssl_certificate_key` di config Nginx menunjuk ke path tersebut (bawaan contoh ini sudah sesuai).

---

## (Alternatif yang Sedang Dipakai) Traefik sebagai Reverse Proxy

Server ini sudah memiliki Traefik (untuk n8n dan layanan lain). Untuk menghindari bentrok port 80/443, kita biarkan Traefik yang melakukan TLS termination dan hanya tambahkan router baru yang meneruskan ke service systemd di port 9300.

### Langkah
1) Pastikan aplikasi jalan di host 0.0.0.0:9300:
   ```bash
   systemctl status wago-api
   curl http://127.0.0.1:9300/health
   ```
2) Buat/replace kontainer router kecil di network Traefik (`root_default`):
   ```bash
   docker rm -f wago-api-router
   docker run -d --name wago-api-router --restart unless-stopped \
     --network root_default \
     -l traefik.enable=true \
     -l "traefik.http.routers.wago-api.rule=Host(`wago-api.chiefaiofficer.id`)" \
     -l traefik.http.routers.wago-api.entrypoints=websecure \
     -l traefik.http.routers.wago-api.tls.certresolver=mytlschallenge \
     -l traefik.http.services.wago-api.loadbalancer.server.url=http://172.18.0.1:9300 \
     traefik/whoami
   ```
   `172.18.0.1` adalah gateway host di network `root_default` (cek dengan `docker network inspect root_default`). Traefik kemudian akan meneruskan ke host:9300.
3) Buka firewall sekali saja agar kontainer Traefik bisa mengakses 9300:
   ```bash
   sudo ufw allow from 172.18.0.0/16 to any port 9300 proto tcp
   ```
4) Verifikasi dari network Traefik:
   ```bash
   docker run --rm --network root_default alpine sh -c 'wget -qO- http://172.18.0.1:9300/health'
   ```
5) Tes publik:
```bash
curl -k https://wago-api.chiefaiofficer.id/health
```
   Harus 200 OK. Sertifikat valid dari Let’s Encrypt via resolver `mytlschallenge` milik Traefik.

### Catatan
- Tidak mengubah konfigurasi atau port Traefik/n8n; hanya menambah router baru.
- Jika gateway network berubah, perbarui URL load balancer dan rule UFW.
- Kontainer `traefik/whoami` hanya sebagai “pemegang label” untuk Traefik; trafik sebenarnya diarahkan ke host:9300.
- Agar router otomatis hidup setelah reboot, buat systemd unit sederhana:
  ```bash
  sudo tee /etc/systemd/system/wago-api-router.service > /dev/null <<'EOF'
  [Unit]
  Description=Wago API Traefik Router
  Requires=docker.service
  After=docker.service

  [Service]
  Type=simple
  Restart=always
  ExecStartPre=-/usr/bin/docker rm -f wago-api-router
  ExecStart=/usr/bin/docker run --name wago-api-router --restart unless-stopped \
    --network root_default \
    -l traefik.enable=true \
    -l traefik.http.routers.wago-api.rule=Host(`wago-api.chiefaiofficer.id`) \
    -l traefik.http.routers.wago-api.entrypoints=websecure \
    -l traefik.http.routers.wago-api.tls.certresolver=mytlschallenge \
    -l traefik.http.services.wago-api.loadbalancer.server.url=http://172.18.0.1:9300 \
    traefik/whoami
  ExecStop=/usr/bin/docker stop wago-api-router

  [Install]
  WantedBy=multi-user.target
  EOF

  sudo systemctl daemon-reload
  sudo systemctl enable --now wago-api-router
  ```
  Jika gateway network berbeda, ubah nilai `172.18.0.1` pada service sebelum mengaktifkan.
