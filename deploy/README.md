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

Kita akan menaruh aplikasi di `/var/www/whatsapp-api`.

```bash
# Buat folder
sudo mkdir -p /var/www/whatsapp-api

# Copy binary dan config
sudo cp whatsapp-api /var/www/whatsapp-api/
sudo cp -r config /var/www/whatsapp-api/
sudo cp -r migrations /var/www/whatsapp-api/
# sudo cp .env /var/www/whatsapp-api/ # (Opsional jika pakai .env)

# Set permission (sesuaikan user, misal: root atau ubuntu)
sudo chown -R root:root /var/www/whatsapp-api
sudo chmod +x /var/www/whatsapp-api/whatsapp-api
```

## 4. Konfigurasi Systemd

Systemd akan menjaga aplikasi tetap berjalan di background dan otomatis restart jika error.

1. Copy file service:
   ```bash
   sudo cp deploy/whatsapp-api.service /etc/systemd/system/
   ```

2. Edit file jika perlu (sesuaikan User, WorkingDirectory, dll):
   ```bash
   sudo nano /etc/systemd/system/whatsapp-api.service
   ```

3. Reload & Start Service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl start whatsapp-api
   sudo systemctl enable whatsapp-api # Agar auto-start saat boot
   ```

4. Cek Status:
   ```bash
   sudo systemctl status whatsapp-api
   ```

## 5. Konfigurasi Nginx (Reverse Proxy)

Nginx akan meneruskan request dari port 80 (HTTP) ke port 8080 (Aplikasi).

1. Buat file config baru di Nginx:
   ```bash
   sudo nano /etc/nginx/sites-available/whatsapp-api
   ```
   *(Copy isi dari file `deploy/nginx.conf` ke sini. Jangan lupa ganti `server_name` dengan domain/IP Anda)*

2. Aktifkan config:
   ```bash
   sudo ln -s /etc/nginx/sites-available/whatsapp-api /etc/nginx/sites-enabled/
   ```

3. Test & Restart Nginx:
   ```bash
   sudo nginx -t
   sudo systemctl restart nginx
   ```

## 6. Selesai!

Sekarang API Anda sudah bisa diakses melalui domain atau IP server tanpa port 8080.
Contoh: `http://api.example.com/api/v1/sessions/create`

---

### Tips Tambahan: SSL (HTTPS)

Untuk mengamankan API dengan HTTPS (gratis via Let's Encrypt):

1. Install Certbot:
   ```bash
   sudo apt install certbot python3-certbot-nginx
   ```

2. Request Sertifikat:
   ```bash
   sudo certbot --nginx -d api.example.com
   ```
