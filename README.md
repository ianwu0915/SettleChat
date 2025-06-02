# SettleChat

SettleChat is a real-time chat system based on WebSocket and NATS, featuring AI-assisted functionality.

## System Architecture

### High-Level System Design

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  WebSocket  │     │   EventBus  │     │    NATS     │
│   Server    │◄───►│    Layer    │◄───►│   Broker    │
└─────────────┘     └─────────────┘     └─────────────┘
       ▲                    ▲                   ▲
       │                    │                   │
       ▼                    ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    Chat     │     │     AI      │     │  Storage    │
│   Module    │     │   Module    │     │   Module    │
└─────────────┘     └─────────────┘     └─────────────┘
```

### HTTP Handlers

#### Authentication Handlers (cmd/server/handler/authHandlers.go)

- **User Registration**

  - Endpoint: `POST /api/auth/register`
  - Functionality: Create new user account
  - Request Body: `{ "username": string, "password": string }`
  - Response: JWT token and user info

- **User Login**
  - Endpoint: `POST /api/auth/login`
  - Functionality: Authenticate user
  - Request Body: `{ "username": string, "password": string }`
  - Response: JWT token and user info

#### Room Handlers (cmd/server/handler/roomHandler.go)

- **Create Room**

  - Endpoint: `POST /api/rooms`
  - Functionality: Create new chat room
  - Request Body: `{ "name": string, "description": string }`
  - Response: Room details

- **Get Room List**

  - Endpoint: `GET /api/rooms`
  - Functionality: List all available rooms
  - Response: Array of room objects

- **Get Room Details**

  - Endpoint: `GET /api/rooms/{roomID}`
  - Functionality: Get specific room information
  - Response: Room details including members

- **Join Room**

  - Endpoint: `POST /api/rooms/{roomID}/join`
  - Functionality: Add user to room
  - Response: Updated room details

- **Leave Room**
  - Endpoint: `POST /api/rooms/{roomID}/leave`
  - Functionality: Remove user from room
  - Response: Success status

#### WebSocket Handler (cmd/server/handler/wshandler.go)

- **WebSocket Connection**
  - Endpoint: `WS /ws`
  - Functionality: Establish WebSocket connection
  - Query Parameters: `roomID`, `token`
  - Features:
    - Real-time message exchange
    - Room state synchronization
    - User presence updates
    - AI command processing

### Message Types

1. **Chat Messages**

   ```json
   {
     "type": "message",
     "content": "string",
     "roomID": "string",
     "senderID": "string",
     "timestamp": "string"
   }
   ```

2. **System Messages**

   ```json
   {
     "type": "system",
     "content": "string",
     "roomID": "string",
     "timestamp": "string"
   }
   ```

3. **AI Commands**
   ```json
   {
     "type": "ai_command",
     "command": "string",
     "content": "string",
     "roomID": "string",
     "senderID": "string",
     "timestamp": "string"
   }
   ```

### Authentication Flow

1. User registers/logs in
2. Server validates credentials
3. JWT token generated
4. Token used for subsequent requests
5. WebSocket connection authenticated using token

### Room Management Flow

1. User creates/joins room
2. Server validates permissions
3. Room state updated
4. Event published to room members
5. WebSocket connections updated

### WebSocket Connection Flow

1. Client initiates connection
2. Server validates token
3. Connection upgraded to WebSocket
4. Client joins room
5. Real-time communication established

### Detailed System Design

#### 1. WebSocket Layer

- Handles user connections and message transmission
- Implements room management and user state tracking
- Provides real-time message push

#### 2. EventBus Layer

- Event routing and distribution
- Message format conversion
- Error handling and retry mechanisms

#### 3. NATS Layer

- Message publish/subscribe
- Topic management
- Message persistence

#### 4. AI Module

- Command processing
- Intelligent response generation
- Context management

#### 5. Storage Module

- Message persistence
- User data management
- Room state maintenance

## Workflows

### Message Transmission Workflow

1. **User Sends Message**

   ```
   Client -> WebSocket -> EventBus -> NATS -> Subscribers
   ```

2. **Message Processing Flow**

   - WebSocket receives message
   - Creates ChatMessage instance
   - Publishes event through EventBus
   - NATS broadcasts to subscribers
   - Storage module persists message

3. **Message Response Flow**
   - Handler generates response
   - Publishes response event through EventBus
   - NATS broadcasts response
   - WebSocket sends to user

### Event Transmission Workflow

1. **Event Publishing**

   ```
   Publisher -> EventBus -> NATS -> Subscribers -> Handlers
   ```

2. **Event Types**

   - Chat message events
   - User state events
   - AI command events
   - System events

3. **Event Processing**
   - Event validation
   - Route distribution
   - Error handling
   - Response generation

## Module Description

### 1. WebSocket Module (internal/chat)

- Client connection management
- Message handling
- Room state synchronization
- User state tracking

### 2. EventBus Module (internal/nats_messaging)

- Event publish/subscribe
- Topic management
- Message format conversion
- Error handling

### 3. AI Module (internal/ai)

- Command handlers
- AI service management
- Context management
- Response generation

### 4. Storage Module (internal/storage)

- Data persistence
- Query optimization
- Transaction management
- Connection pool management

## Setup and Running

### Prerequisites

- Docker
- Go 1.21 or later
- Make (optional)

### Step 1: Start Required Services

1. **Start NATS Server**

```bash
docker run -d --name nats-server -p 4222:4222 nats:latest
```

2. **Start PostgreSQL**

```bash
docker run -d --name postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=password \
    -e POSTGRES_DB=settlechat \
    -p 5432:5432 \
    postgres:latest
```

### Step 2: Configure Environment

Create a `.env` file in the project root:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=settlechat
NATS_URL=nats://localhost:4222
WS_PORT=8080
```

### Step 3: Run the Application

1. **Install Dependencies**

```bash
go mod download
```

2. **Run the Server**

```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default.

### Step 4: Access the Chat

Open your browser and navigate to:

```
http://localhost:8080
```

## Configuration

### Environment Variables

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=settlechat
NATS_URL=nats://localhost:4222
WS_PORT=8080
```

### Topic Format

```
settlechat{env}.{category}.{action}.{roomID}
```

## Development Guide

### Local Development

1. Clone the repository
2. Install dependencies
3. Configure environment variables
4. Run the service

### Testing

```bash
go test ./...
```

### Deployment

1. Build Docker image
2. Configure environment variables
3. Deploy service

## Monitoring and Logging

### Log Levels

- DEBUG: Detailed debugging information
- INFO: General operational information
- WARN: Warning messages
- ERROR: Error messages

### Monitoring Metrics

- Connection count
- Message throughput
- Response time
- Error rate

## Error Handling

### Error Types

1. Connection errors
2. Message processing errors
3. AI processing errors
4. Storage errors

### Error Recovery

- Automatic reconnection
- Message retry
- Error logging
- User notification

## Performance Optimization

### Optimization Strategies

1. Connection pool management
2. Message batching
3. Caching mechanisms
4. Asynchronous processing

### Monitoring Metrics

1. Response time
2. Throughput
3. Resource utilization
4. Error rate

## Security Considerations

### Security Measures

1. Authentication
2. Message encryption
3. Access control
4. Input validation

### Best Practices

1. Regular updates
2. Security audits
3. Log monitoring
4. Backup strategies

## Project Structure

```
settlechat/
│
├── cmd/
│   └── server/             # Entry point main.go
│       └── main.go
│
├── internal/               # Core logic, not exposed externally
│   ├── ws/                 # WebSocket handler, connection upgrade, entry point
│   ├── chat/               # Chatroom, Client, Hub core logic
│   ├── ai/                 # AI integration (DeepSeek, OpenAI, etc.)
│   ├── storage/            # Message persistence logic (currently using SQLite)
│   ├── command/            # Logic for handling commands like `/summary`
│   └── utils/              # UUID, time, string processing
│
├── web/                    # Static resources (HTML/JS/CSS) or frontend build output
│   └── index.html
│
├── go.mod
├── go.sum
└── README.md
```

## Future Scaling Considerations

1. Client-to-Server Assignment => Load Balancing + Consistent Hashing
2. Read/Write Optimization with Database Choices
3. Middleware Like Kafka/Flink for High Message Throughput
4. Caching Implementation
5. Distributed ID Generation
6. Persistent Client Connections with Zookeeper and Load Balancer
7. Message Streaming to Groups Before Database Storage using Flink or Kafka
8. Redesign Using Single WebSocket Connection per Server

### Setup-NATS Server

```bash
docker run -d --name nats-server -p 4222:4222 nats:latest
```

### AI-PART Desgin
