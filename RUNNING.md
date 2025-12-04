# How to Run (No Docker)

## Prerequisites
1. Go 1.21+
2. PostgreSQL 14+

## Setup

1. **Database Setup**
   Ensure PostgreSQL is running. Create a database named `whatsapp_api`.
   ```bash
   createdb -U postgres whatsapp_api
   ```
   *Note: Adjust credentials in `.env` if necessary.*

2. **Migrations**
   Run the SQL scripts in `migrations/` folder to create tables.
   You can run them manually using `psql`:
   ```bash
   psql -U postgres -d whatsapp_api -f migrations/001_create_users_table.up.sql
   psql -U postgres -d whatsapp_api -f migrations/002_create_sessions_table.up.sql
   psql -U postgres -d whatsapp_api -f migrations/003_create_messages_table.up.sql
   psql -U postgres -d whatsapp_api -f migrations/004_create_langchain_executions_table.up.sql
   ```

3. **Configuration**
   Check `config/config.yaml` and `.env` to match your local environment.

## Running the API

### Option 1: Run Directly (Development)
   Useful for development as it compiles and runs in one step.
   ```bash
   go run cmd/api/main.go
   ```

### Option 2: Build and Run (Production)
   Better for production as it produces a binary file.
   
   1. **Build**
      ```bash
      go build -o whatsapp-api cmd/api/main.go
      ```

   2. **Run**
      ```bash
      ./whatsapp-api
      ```

## Usage

The server will start on port 8080.
A default user is created automatically if none exists:
- **UserID**: `admin`
- **API Key**: `secret`

### Create Session
```bash
curl -X POST http://localhost:8080/api/v1/sessions/create \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"agentId": "agent_01", "agentName": "My Bot"}'
```

### Swagger Documentation
You can access the Swagger UI at:
http://localhost:8080/swagger/index.html

Use the API Key `secret` in the "Authorize" button (value: `Bearer secret`).
