# Product Requirements Document (PRD)
## WhatsApp REST API with Go

---

## 1. Executive Summary

### 1.1 Product Overview
Aplikasi REST API WhatsApp berbasis Go yang cepat, ringan, dan scalable untuk mendukung multi-user. API ini menyediakan kemampuan untuk mengelola sesi WhatsApp, mengirim pesan, dan mengeksekusi Langchain API dengan database PostgreSQL sebagai storage utama.

### 1.2 Goals
- Membangun REST API yang cepat dan ringan menggunakan Go
- Mendukung multi-user dengan session management yang robust
- Implementasi clean architecture untuk maintainability
- Integrasi dengan PostgreSQL untuk persistence
- Generate dan manage QR Code realtime untuk WhatsApp authentication
- Integrasi dengan Langchain untuk AI capabilities

---

## 2. Technical Stack

### 2.1 Core Technologies
- **Language**: Go 1.21+
- **Database**: PostgreSQL 14+
- **WhatsApp Library**: `go.mau.fi/whatsmeow`
- **Web Framework**: `github.com/gofiber/fiber/v2` (fast & lightweight)
- **Database Driver**: `github.com/jackc/pgx/v5`
- **ORM/Query Builder**: `github.com/jmoiron/sqlx` atau raw SQL
- **QR Code**: `github.com/skip2/go-qrcode`
- **Configuration**: `github.com/spf13/viper`
- **Logging**: `github.com/sirupsen/logrus` atau `go.uber.org/zap`

### 2.2 Project Structure (Clean Architecture)
```
whatsapp-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── session.go
│   │   │   ├── message.go
│   │   │   └── user.go
│   │   └── repository/
│   │       ├── session_repository.go
│   │       └── message_repository.go
│   ├── usecase/
│   │   ├── session_usecase.go
│   │   ├── message_usecase.go
│   │   └── langchain_usecase.go
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── session_handler.go
│   │       │   ├── message_handler.go
│   │       │   └── langchain_handler.go
│   │       ├── middleware/
│   │       │   ├── auth.go
│   │       │   └── logger.go
│   │       └── router.go
│   └── infrastructure/
│       ├── database/
│       │   ├── postgres.go
│       │   └── migration.go
│       ├── whatsapp/
│       │   ├── client.go
│       │   └── qr_generator.go
│       └── langchain/
│           └── client.go
├── pkg/
│   ├── config/
│   │   └── config.go
│   ├── logger/
│   │   └── logger.go
│   └── utils/
│       └── response.go
├── migrations/
│   ├── 001_create_users_table.up.sql
│   ├── 001_create_users_table.down.sql
│   ├── 002_create_sessions_table.up.sql
│   ├── 002_create_sessions_table.down.sql
│   ├── 003_create_messages_table.up.sql
│   └── 003_create_messages_table.down.sql
├── config/
│   └── config.yaml
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## 3. Database Schema (PostgreSQL)

### 3.1 Table: users
```sql
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_api_key ON users(api_key);
```

### 3.2 Table: sessions
```sql
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    agent_name VARCHAR(255),
    phone_number VARCHAR(50),
    qr_code TEXT,
    qr_code_base64 TEXT,
    session_data JSONB,
    status VARCHAR(50) DEFAULT 'disconnected', -- connected, disconnected, expired
    langchain_url TEXT,
    last_qr_generated_at TIMESTAMP,
    connected_at TIMESTAMP,
    disconnected_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, agent_id)
);

CREATE INDEX idx_sessions_user_agent ON sessions(user_id, agent_id);
CREATE INDEX idx_sessions_status ON sessions(status);
```

### 3.3 Table: messages
```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(255),
    from_number VARCHAR(50),
    to_number VARCHAR(50),
    message_text TEXT,
    message_type VARCHAR(50), -- text, image, document, etc
    direction VARCHAR(20), -- incoming, outgoing
    status VARCHAR(50), -- sent, delivered, read, failed
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_session ON messages(session_id);
CREATE INDEX idx_messages_agent ON messages(agent_id);
CREATE INDEX idx_messages_created ON messages(created_at DESC);
```

### 3.4 Table: langchain_executions
```sql
CREATE TABLE langchain_executions (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    user_message TEXT,
    langchain_response JSONB,
    execution_time_ms INTEGER,
    status VARCHAR(50), -- success, failed
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_langchain_session ON langchain_executions(session_id);
CREATE INDEX idx_langchain_created ON langchain_executions(created_at DESC);
```

---

## 4. API Endpoints

### 4.1 Session Management

#### POST /api/v1/sessions/create
**Description**: Create new WhatsApp session and generate QR code

**Headers**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**Request Body**:
```json
{
  "userId": "user_123",
  "agentId": "agent_456",
  "agentName": "Customer Service Bot"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Session created successfully",
  "data": {
    "sessionId": 1,
    "agentId": "agent_456",
    "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANS...",
    "status": "waiting_scan",
    "expiresIn": 60
  }
}
```

---

#### GET /api/v1/sessions/status
**Description**: Get session status (connected/disconnected)

**Headers**:
```
Authorization: Bearer {API_KEY}
```

**Request Body**:
```json
{
  "agentId": "agent_456"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "agentId": "agent_456",
    "status": "connected",
    "phoneNumber": "628123456789",
    "connectedAt": "2024-01-15T10:30:00Z"
  }
}
```

---

#### PUT /api/v1/sessions/reconnect
**Description**: Reconnect session (delete old, create new with fresh QR)

**Headers**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**Request Body**:
```json
{
  "agentId": "agent_456"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Session reconnected, scan new QR code",
  "data": {
    "sessionId": 2,
    "agentId": "agent_456",
    "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANS...",
    "status": "waiting_scan"
  }
}
```

---

#### DELETE /api/v1/sessions/delete
**Description**: Delete session

**Headers**:
```
Authorization: Bearer {API_KEY}
```

**Request Body**:
```json
{
  "agentId": "agent_456"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Session deleted successfully"
}
```

---

#### GET /api/v1/sessions/detail
**Description**: Get detailed session information

**Headers**:
```
Authorization: Bearer {API_KEY}
```

**Request Body**:
```json
{
  "agentId": "agent_456"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "sessionId": 1,
    "agentId": "agent_456",
    "agentName": "Customer Service Bot",
    "phoneNumber": "628123456789",
    "status": "connected",
    "connectedAt": "2024-01-15T10:30:00Z",
    "incomingMessages": 150,
    "outgoingMessages": 145,
    "langchainExecutions": 50,
    "lastActivity": "2024-01-15T14:25:00Z"
  }
}
```

---

### 4.2 Langchain Execution

#### POST /api/v1/langchain/execute
**Description**: Execute Langchain API

**Headers**:
```
Authorization: Bearer {API_KEY}
Content-Type: application/json
```

**Request Body**:
```json
{
  "userId": "user_123",
  "agentId": "agent_456",
  "apiKey": "langchain_api_key",
  "langchainUrl": "https://new-langchain.chiefaiofficer.id/v1/api",
  "input": {
    "message": "User message here",
    "parameters": {
      "max_steps": 5
    }
  },
  "sessionId": "628123456789"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Langchain executed successfully",
  "data": {
    "executionId": 1,
    "langchainResponse": {
      "output": "AI response here",
      "steps": 3,
      "metadata": {}
    },
    "executionTimeMs": 1250
  }
}
```

---

### 4.3 Message Management (Optional/Future Enhancement)

#### POST /api/v1/messages/send
**Description**: Send WhatsApp message

**Request Body**:
```json
{
  "agentId": "agent_456",
  "to": "628123456789",
  "message": "Hello from API",
  "type": "text"
}
```

---

## 5. Core Features & Implementation

### 5.1 Multi-User Support
- Setiap user memiliki API key unik
- Isolasi data berdasarkan user_id
- Session management per user dan agent
- Rate limiting per user

### 5.2 QR Code Management
- Auto-generate QR code dalam format Base64
- Auto-refresh QR jika expired (60 detik)
- Notifikasi real-time ketika QR di-scan
- Store QR code di database untuk audit

### 5.3 Session Management
- Persistent storage di PostgreSQL
- Auto-reconnect mechanism
- Session health check
- Graceful disconnect handling

### 5.4 Performance Optimization
- Connection pooling untuk database
- Goroutines untuk concurrent operations
- Caching untuk frequently accessed data (Redis optional)
- Efficient JSON serialization

### 5.5 Error Handling & Logging
- Structured logging dengan contextual information
- Error tracking dan monitoring
- Graceful degradation
- Detailed error responses

### 5.6 Security
- API key authentication
- Rate limiting per user
- Input validation & sanitization
- SQL injection prevention
- HTTPS enforcement

---

## 6. Configuration

### 6.1 Environment Variables (.env)
```env
# Server
APP_ENV=production
APP_PORT=8080
APP_NAME=WhatsApp-API

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=whatsapp_api
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5

# WhatsApp
WA_AUTO_RECONNECT=true
WA_QR_TIMEOUT=60
WA_LOG_LEVEL=INFO

# Langchain
LANGCHAIN_DEFAULT_TIMEOUT=30s
LANGCHAIN_MAX_RETRIES=3

# Security
API_KEY_HEADER=Authorization
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 7. API Response Format

### 7.1 Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    // response data
  }
}
```

### 7.2 Error Response
```json
{
  "success": false,
  "error": {
    "code": "INVALID_SESSION",
    "message": "Session not found or expired",
    "details": {
      "agentId": "agent_456"
    }
  }
}
```

---

## 8. Error Codes

| Code | Description |
|------|-------------|
| `INVALID_API_KEY` | API key tidak valid atau expired |
| `SESSION_NOT_FOUND` | Session tidak ditemukan |
| `SESSION_EXPIRED` | Session telah expired |
| `QR_GENERATION_FAILED` | Gagal generate QR code |
| `WHATSAPP_CONNECTION_FAILED` | Gagal connect ke WhatsApp |
| `LANGCHAIN_EXECUTION_FAILED` | Gagal eksekusi Langchain |
| `DATABASE_ERROR` | Database operation error |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `INVALID_REQUEST` | Request body tidak valid |

---

## 9. Performance Requirements

### 9.1 Response Time
- API endpoint: < 200ms (average)
- QR generation: < 500ms
- Langchain execution: < 5s (tergantung Langchain)

### 9.2 Throughput
- Support 1000+ concurrent connections
- Handle 10,000+ requests/minute

### 9.3 Availability
- 99.9% uptime
- Auto-restart on failure
- Health check endpoint

---

## 10. Monitoring & Observability

### 10.1 Health Check Endpoint
```
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "components": {
    "database": "healthy",
    "whatsapp": "healthy"
  }
}
```

### 10.2 Metrics to Track
- API response time
- Error rate per endpoint
- Active sessions count
- Database connection pool status
- Message throughput

---

## 11. Deployment

### 11.1 Docker Support
- Dockerfile untuk containerization
- docker-compose.yml untuk local development
- Multi-stage build untuk production

### 11.2 Database Migration
- Auto-run migrations on startup
- Rollback support
- Migration versioning

---

## 12. Testing Strategy

### 12.1 Unit Tests
- Repository layer
- Use case layer
- Utility functions

### 12.2 Integration Tests
- API endpoints
- Database operations
- WhatsApp client

### 12.3 Load Testing
- Concurrent sessions
- High throughput scenarios

---

## 13. Documentation

### 13.1 Required Documentation
- API documentation (Swagger/OpenAPI)
- Setup guide (README.md)
- Architecture diagram
- Database schema documentation
- Deployment guide

### 13.2 Code Documentation
- Inline comments untuk complex logic
- Function/method documentation
- Package documentation

---

## 14. Development Phases

### Phase 1: Foundation (Week 1)
- [ ] Project setup & structure
- [ ] Database schema & migrations
- [ ] Configuration management
- [ ] Basic HTTP server setup

### Phase 2: Core Features (Week 2-3)
- [ ] Session management endpoints
- [ ] WhatsApp client integration
- [ ] QR code generation
- [ ] Database repositories

### Phase 3: Advanced Features (Week 4)
- [ ] Langchain integration
- [ ] Message handling
- [ ] Error handling & logging
- [ ] API key authentication

### Phase 4: Optimization & Testing (Week 5)
- [ ] Performance optimization
- [ ] Unit & integration tests
- [ ] Load testing
- [ ] Documentation

### Phase 5: Deployment (Week 6)
- [ ] Docker setup
- [ ] CI/CD pipeline
- [ ] Production deployment
- [ ] Monitoring setup

---

## 15. Success Criteria

- ✅ API response time < 200ms untuk 95% requests
- ✅ Support minimal 100 concurrent users
- ✅ Code coverage > 80%
- ✅ Zero critical security vulnerabilities
- ✅ Complete API documentation
- ✅ Successful load testing with 1000 concurrent sessions
- ✅ Clean code dengan Go best practices

---

## 16. Future Enhancements

- WebSocket support untuk real-time updates
- Redis caching layer
- Message queue (RabbitMQ/Kafka) untuk async processing
- Webhook support untuk incoming messages
- Multi-device support
- Admin dashboard
- Analytics & reporting

---

**Document Version**: 1.0  
**Last Updated**: 2024-01-15  
**Owner**: Engineering Team