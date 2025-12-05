# Deploy Log

## 2025-12-05
- Added .env loading + env bindings in `pkg/config` so SERVER_PORT/DATABASE_URL envs override config.
- Created `.env` with remote PostgreSQL DSN and adjusted Go app to probe for free ports when binding.
- Ran remote DB migrations to provision schema and seeded default user.
- Built binary and deployed to `/var/www/wago-api` with systemd unit `wago-api.service` targeting port 9300.
- Configured Nginx (`/etc/nginx/sites-available/wago-api`) to terminate TLS for `wago-api.chiefaiofficer.id`, proxy to `127.0.0.1:9300`, and log to `wago-api-*` files.
- Issued valid Let's Encrypt cert via `certbot --nginx -d wago-api.chiefaiofficer.id`.
- Verified service via `systemctl status wago-api` and HTTPS swagger access.
