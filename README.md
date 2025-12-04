# Whatsapp API (Go)

Self-hosted WhatsApp automation API built with Go, Fiber, and WhatsMeow. It manages session lifecycle, proxies messages to LangChain agents, and exposes REST endpoints plus generated Swagger docs.

## Features
- WhatsApp session management (QR, reconnect, status)
- Message ingest + optional persistence
- LangChain integration for AI replies
- Fiber HTTP API with Swagger UI
- SQL migrations for PostgreSQL

## Architecture at a Glance
- Entrypoint: `cmd/api/main.go` wires config, DB, use cases, and HTTP routes.
- Layers: `internal/delivery/http` (routes/handlers), `internal/usecase` (business logic), `internal/domain` (entities/repo interfaces), `internal/infrastructure` (DB, WhatsApp client, LangChain).
- Config: `pkg/config` loads `config/config.yaml` with env overrides (`.env`).
- Docs: `docs/` served at `/swagger/index.html` (regenerate via `swag init -g cmd/api/main.go -o docs`).
- Migrations: `migrations/*.sql` target PostgreSQL DB `whatsapp_api`.

## Requirements
- Go 1.24+
- PostgreSQL (local or remote)
- WhatsApp account for device pairing

## Setup
1) Copy `.env.example` (if available) to `.env` and fill DB credentials, LangChain URL/API key, default user, and API key for HTTP auth.  
2) Adjust non-secret defaults in `config/config.yaml` as needed (kept out of VCS).  
3) Apply migrations in order:
```bash
psql -U postgres -d whatsapp_api -f migrations/001_init.up.sql
# ...continue with next numbered migrations
```

## Run (Dev)
```bash
go run cmd/api/main.go
```
The service reads `.env` and `config/config.yaml`, emits a QR in logs/DB, and listens for API requests. Swagger UI: `http://localhost:3000/swagger/index.html`.

## Build
```bash
go build -o whatsapp-api cmd/api/main.go
./whatsapp-api
```

## Tests
```bash
go test ./...
```
Add table-driven unit tests alongside new code paths.

## Regenerate Swagger
After handler/route changes:
```bash
swag init -g cmd/api/main.go -o docs
```

## Development Conventions
- Use `gofmt` (tabs) and `go vet` before commits.
- Exported types/interfaces in PascalCase; helpers in lowerCamelCase.
- Handlers return `error` for Fiber to handle.
- Avoid committing secrets; keep `.env` local and rotate default API keys outside dev.

## Common Endpoints (high level)
- Session lifecycle (create, reconnect, delete)
- Message send/receive hooks with LangChain execution
- Health/status endpoints

Refer to Swagger for exact paths and payloads.

## Troubleshooting
- **QR not refreshing**: ensure DB reachable and check `LastQRGeneratedAt`; reconnect triggers fresh QR.
- **No replies in group**: confirm bot is mentioned (by number or LID) and LangChain URL/API key are set.
- **DB errors**: verify DSN in `.env` and that migrations ran.
