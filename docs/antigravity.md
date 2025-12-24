# Project Study: Whatsapp-API-GO

## Overview
**Whatsapp-API-GO** is a high-performance, scalable REST API built with Go, designed to manage WhatsApp sessions and integrate with AI agents via Langchain. It serves as a bridge between WhatsApp users and AI capabilities, offering robust session management and message handling.

## Technical Stack
- **Language**: Go 1.24
- **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber) (Fast & Lightweight)
- **Database**: PostgreSQL (using `pgx` driver)
- **WhatsApp Library**: [whatsmeow](https://go.mau.fi/whatsmeow)
- **AI Integration**: Custom Langchain Client
- **Configuration**: Viper
- **Documentation**: Swagger/OpenAPI

## Architecture
The project follows **Clean Architecture** principles to ensure maintainability and testability:

### Directory Structure
- **`cmd/api/`**: Application entry point (`main.go`). Initializes dependencies and starts the server.
- **`internal/`**: Core application logic.
    - **`domain/`**: Entities and repository interfaces.
    - **`usecase/`**: Business logic (Session, Message, Langchain).
    - **`delivery/http/`**: HTTP handlers and routers (Fiber).
    - **`infrastructure/`**: External implementations (Database, WhatsApp, Langchain).
- **`migrations/`**: Database schema migrations.
- **`pkg/`**: Shared utilities (Config, Logger).

## Key Features

### 1. Session Management
- **Multi-User Support**: Manages multiple WhatsApp sessions concurrently.
- **QR Code Generation**: Generates QR codes for authentication.
- **State Management**: Tracks session status (connected, disconnected, waiting_scan).
- **Endpoints**:
    - `POST /api/v1/sessions/create`: Create a new session.
    - `GET /api/v1/sessions/status`: Check connection status.
    - `GET /api/v1/sessions/detail`: Get detailed session info.
    - `POST /api/v1/sessions/reconnect`: Reconnect a session.
    - `DELETE /api/v1/sessions/delete`: Remove a session.

### 2. AI Integration (Langchain)
- **Client**: Located in `internal/infrastructure/langchain/client.go`.
- **Functionality**: Sends user messages to an external Langchain API and retrieves AI responses.
- **Endpoint**: `POST /api/v1/langchain/execute`.
- **Configuration**: Supports dynamic Base URL, Agent ID, and API Key per request.

### 3. Database Schema
- **`users`**: Stores API keys and user info.
- **`sessions`**: Stores WhatsApp session data (QR, status, agent info).
- **`messages`**: Logs incoming and outgoing messages.
- **`langchain_executions`**: Tracks AI interactions and performance.

## Application Flow
1.  **Initialization**: `main.go` loads config, connects to Postgres, runs migrations, and initializes the WhatsApp client manager.
2.  **Routing**: `internal/delivery/http/router.go` maps HTTP requests to specific handlers.
3.  **Request Handling**: Handlers invoke UseCases to perform business logic.
4.  **Infrastructure**: UseCases interact with Repositories (DB) and External Services (WhatsApp, Langchain).

## Observations
- The project is production-ready with proper error handling, logging, and configuration.
- It uses `context` for timeout management, especially in external API calls.
- The modular design allows for easy extension, such as adding new message handlers or AI providers.

## API Endpoints

### 1. Session Management

#### `POST /api/v1/sessions/create`
Creates a new WhatsApp session and generates a QR code for authentication.
- **Request Body**:
    ```json
    {
        "agentId": "string (required)",
        "agentName": "string",
        "apiKey": "string",
        "langchainUrl": "string"
    }
    ```
- **Response** (200 OK):
    ```json
    {
        "success": true,
        "message": "Session created successfully",
        "data": {
            "sessionId": 1,
            "agentId": "agent_123",
            "qrCode": "data:image/png;base64,...",
            "qrCodeBase64": "...",
            "status": "waiting_scan",
            "lastGeneratedAt": "2024-01-01T00:00:00Z"
        }
    }
    ```

#### `GET /api/v1/sessions/status`
Retrieves the current status of a session.
- **Query Parameter**: `agentId` (string, required)
- **Response** (200 OK):
    ```json
    {
        "success": true,
        "data": {
            "agentId": "agent_123",
            "status": "connected",
            "phoneNumber": "628123456789",
            "connectedAt": "2024-01-01T00:00:00Z",
            "qrCode": "...",
            "lastQrGeneratedAt": "..."
        }
    }
    ```

#### `GET /api/v1/sessions/detail`
Retrieves detailed information about a session, including message statistics.
- **Query Parameter**: `agentId` (string, required)
- **Response** (200 OK):
    ```json
    {
        "success": true,
        "data": {
            "session": { ... }, // Full session object
            "stats": {
                "incoming": 10,
                "responded": 8
            }
        }
    }
    ```

#### `POST /api/v1/sessions/reconnect`
Forces a session reconnection, generating a new QR code if necessary.
- **Request Body**:
    ```json
    {
        "agentId": "string (required)"
    }
    ```
- **Response** (200 OK):
    ```json
    {
        "success": true,
        "message": "Session reconnected successfully",
        "data": { ... } // Similar to create session data
    }
    ```

#### `DELETE /api/v1/sessions/delete`
Deletes an existing session.
- **Request Body**:
    ```json
    {
        "agentId": "string (required)"
    }
    ```
- **Response** (200 OK):
    ```json
    {
        "success": true,
        "message": "Session deleted successfully"
    }
    ```

### 2. AI Integration

#### `POST /api/v1/langchain/execute`
Proxies a message to the configured Langchain agent and stores the execution result.
- **Request Body**:
    ```json
    {
        "agentId": "string (required)",
        "message": "string (required)",
        "sender": "string (optional)",
        "params": { ... } // Optional parameters for Langchain
    }
    ```
- **Response** (200 OK):
    ```json
    {
        "success": true,
        "data": {
            "id": 1,
            "agentId": "agent_123",
            "status": "success",
            "userMessage": "Hello",
            "langchainResponse": { ... },
            "executionTimeMs": 150,
            "createdAt": "..."
        }
    }
    ```

### 3. System

#### `GET /health`
Health check endpoint.
- **Response** (200 OK):
    ```json
    {
        "status": "ok"
    }
    ```
