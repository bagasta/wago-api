# Deployment with Docker + Nginx

Instruksi berikut menyiapkan aplikasi beserta PostgreSQL dengan Docker Compose dan meneruskannya lewat Nginx ke domain `wago-api.chiefaiofficer.id`. Sesuaikan port jika ada service lain yang sedang berjalan.

## 1) Persiapan
- Pastikan Docker dan Docker Compose terinstal.
- Siapkan DNS `A` record `wago-api.chiefaiofficer.id` yang mengarah ke server.
- Salin konfigurasi env: `cp .env.example .env` lalu sesuaikan nilai (ganti password DB dan port jika bentrok). Default port aplikasi: `9300`; port host PostgreSQL: `5544`.

## 2) Jalankan stack Docker (pakai Postgres lokal)
Pastikan Postgres lokal Anda sudah berjalan dan dapat diakses dari host (misal di port 5432). Sesuaikan URL di `.env` jika port berbeda.
```bash
docker compose build
docker compose up -d
docker compose ps
```
- Aplikasi tersedia di `http://127.0.0.1:${APP_PORT:-9300}` pada host. Jika port sudah dipakai, ubah `APP_PORT` di `.env` lalu `docker compose up -d`.
- Koneksi DB di dalam kontainer diarahkan ke `host.docker.internal` (Linux memakai `extra_hosts`), sehingga kontainer bisa mengakses Postgres lokal host.

## 3) Konfigurasi Nginx (host)
Gunakan `deploy/nginx.conf` sebagai template dan arahkan proxy ke port aplikasi di host (default `9300`):
```nginx
location / {
    proxy_pass http://127.0.0.1:9300;
    # header forward lain tetap sama
}
```
- Letakkan file di `/etc/nginx/sites-available/wago-api` lalu symlink ke `sites-enabled`.
- Pastikan sertifikat Let’s Encrypt tersedia di `/etc/letsencrypt/live/wago-api.chiefaiofficer.id/` (lihat langkah SSL).
- Tes konfigurasi `sudo nginx -t` lalu `sudo systemctl reload nginx`.

## 4) Sertifikat SSL
Jika sertifikat belum ada, jalankan (setelah DNS mengarah):
```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d wago-api.chiefaiofficer.id
```
Pastikan `ssl_certificate` dan `ssl_certificate_key` di `nginx.conf` menunjuk ke path yang dibuat Certbot.

## 5) Operasional
- Lihat log: `docker compose logs -f app` atau `docker compose logs -f db`.
- Hentikan stack: `docker compose down` (data DB tetap tersimpan di volume `wago_api_db_data`).
- Perbarui image setelah perubahan kode: `docker compose build --no-cache` lalu `docker compose up -d`.
