# Deploy Log

## 2025-12-05
- Added .env loading + env bindings in `pkg/config` so SERVER_PORT/DATABASE_URL envs override config.
- Created `.env` with remote PostgreSQL DSN and adjusted Go app to probe for free ports when binding.
- Ran remote DB migrations to provision schema and seeded default user.
- Built binary and deployed to `/var/www/wago-api` with systemd unit `wago-api.service` targeting port 9300.
- Configured Nginx (`/etc/nginx/sites-available/wago-api`) to terminate TLS for `wago-api.chiefaiofficer.id`, proxy to `127.0.0.1:9300`, and log to `wago-api-*` files.
- Issued valid Let's Encrypt cert via `certbot --nginx -d wago-api.chiefaiofficer.id`.
- Verified service via `systemctl status wago-api` and HTTPS swagger access.

## 2025-12-09
- Shifted public entrypoint to Traefik (existing stack) instead of host-level Nginx to avoid port conflicts.
- Deployed lightweight router container `wago-api-router` on network `root_default` with labels:
  - `traefik.http.routers.wago-api.rule=Host(\`wago-api.chiefaiofficer.id\`)`
  - `traefik.http.routers.wago-api.entrypoints=websecure`
  - `traefik.http.routers.wago-api.tls.certresolver=mytlschallenge`
  - `traefik.http.services.wago-api.loadbalancer.server.url=http://172.18.0.1:9300`
- Opened UFW to allow Docker bridge traffic to app: `ufw allow from 172.18.0.0/16 to any port 9300 proto tcp` (retained 172.17.0.0/16 rule).
- Verified health from Traefik network and external HTTPS: `https://wago-api.chiefaiofficer.id/health` returns 200 with valid Let's Encrypt cert.
- Added systemd unit `/etc/systemd/system/wago-api-router.service` to ensure Traefik router container (`wago-api-router`) starts on boot and carries required labels for routing.
