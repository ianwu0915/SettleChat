# SettleChat

> **AI-Driven Real-Time Collaboration Platform**  
> Intelligent chat system with event-driven architecture, WebSocket communication, and AI-powered conversation assistance.

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![NATS](https://img.shields.io/badge/messaging-NATS-green.svg)](https://nats.io)
[![PostgreSQL](https://img.shields.io/badge/database-PostgreSQL-blue.svg)](https://postgresql.org)
[![Docker](https://img.shields.io/badge/deployment-Docker-blue.svg)](https://docker.com)

## 🚀 Overview

SettleChat is a modern, scalable real-time chat application that combines traditional messaging with AI-powered collaboration features. Built with Go's performance and NATS' distributed messaging capabilities, it provides intelligent conversation assistance, automatic summarization, and seamless team collaboration tools.

### ✨ Key Features

- **🔄 Real-Time Communication**: WebSocket-based messaging with automatic reconnection
- **🤖 AI-Powered Assistance**: Intelligent conversation summaries, command processing, and collaboration insights
- **📡 Event-Driven Architecture**: NATS-based message distribution for high scalability
- **🏠 Multi-Room Support**: Create and manage multiple chat rooms with persistent history
- **👥 User Management**: Secure authentication with bcrypt encryption
- **💾 Data Persistence**: PostgreSQL storage with optimized indexing for chat history
- **📈 Performance Monitoring**: Built-in benchmarking and performance testing
- **🐳 Container Ready**: Full Docker Compose setup for easy deployment

## 🏗️ Architecture

### System Design

```
┌─────────────────┐
│  Frontend       │
│  (WebSocket)    │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│  Chat Module    │
│ (Hub/Room/Client)│
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   EventBus      │
│  (Abstraction)  │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│     NATS        │
│  (Messaging)    │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Event Handlers  │
│ (Business Logic)│
└───────┬─────────┘
        │
    ┌───▼───┐
    │       │
    ▼       ▼
┌─────┐   ┌─────────┐
│ AI  │   │ Storage │
│Module│   │ Module  │
└─────┘   └─────────┘
```

### Layered Architecture

#### 🎯 EventBus Layer (Abstraction)

- **Purpose**: Unified event publishing interface that abstracts underlying messaging system
- **Location**: `internal/messaging/eventbus.go`
- **Key Methods**:
  - `PublishEvent()` - Routes events to appropriate NATS topics
  - `PublishUserJoinedEvent()`, `PublishAICommandEvent()` - Typed event publishers
- **Benefits**: Decouples business logic from NATS, enables easy testing and message system swapping

#### 🚀 NATS Layer (Implementation)

- **Purpose**: High-performance distributed messaging and event streaming
- **Location**: `internal/messaging/nats/`
- **Components**:
  - `NATSManager` - Connection lifecycle and reliability
  - `Publisher/Subscriber` - Message distribution
  - `TopicFormatter` - Consistent topic naming (`settlechat.{category}.{action}.{roomID}`)

#### 💬 Core Components

- **Hub & Room Management**: Central WebSocket connection management with room-based message routing
- **AI Integration**: Pluggable AI providers (DeepSeek, OpenAI) with intelligent command processing
- **Storage Layer**: PostgreSQL with optimized indexing for chat history and user management
- **Event Handlers**: Specialized processors for different event types (user actions, AI commands, etc.)

## 🛠️ Technology Stack

| Layer              | Technology          | Purpose                             |
| ------------------ | ------------------- | ----------------------------------- |
| **Backend**        | Go 1.24+            | High-performance server runtime     |
| **Event System**   | EventBus + NATS     | Abstracted event-driven messaging   |
| **Database**       | PostgreSQL          | Persistent data storage             |
| **Real-time**      | WebSocket           | Bidirectional client communication  |
| **AI**             | DeepSeek/OpenAI     | Intelligent conversation processing |
| **Infrastructure** | Docker Compose      | Service orchestration               |
| **Frontend**       | Vanilla JS/HTML/CSS | Lightweight client interface        |

### Message Flow Architecture

```
📱 User Input (Frontend)
          ↓
🔌 WebSocket Connection (client.go)
          ↓
🏠 Chat Module (Hub/Room/Client) - Connection & Room Management
          ↓
📤 EventBus.PublishEvent() - Event Abstraction Layer
          ↓
🚀 NATS Topic - Distributed Messaging
          ↓
📥 Event Handlers - Business Logic Processing
          ↓
    ┌─────────┴─────────┐
    ▼                   ▼
🤖 AI Module        💾 Storage Module
    │                   │
    └─────────┬─────────┘
              ▼
📤 EventBus.PublishResponse() - Response Events
              ▼
🚀 NATS Broadcast - Message Distribution
              ▼
🔌 WebSocket WritePump - Real-time Delivery
              ▼
📱 Frontend Update (All Connected Clients)

Key Flow Steps:
1. Frontend sends message via WebSocket
2. Chat Module (Client) receives and routes
3. EventBus abstracts event publishing to NATS
4. NATS distributes to appropriate Event Handlers
5. Handlers process business logic (AI/Storage)
6. Response events published back through EventBus
7. Real-time updates delivered to all clients
```

### Event-Driven Benefits

- **🔄 Loose Coupling**: EventBus abstracts NATS, making system components independent
- **📈 Scalability**: NATS enables horizontal scaling and load distribution
- **🧪 Testability**: Easy to mock EventBus for unit testing
- **🔄 Flexibility**: Can replace NATS with other message systems (Kafka, Redis) without changing business logic
- **⚡ Performance**: Asynchronous event processing prevents blocking operations

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+**
- **Docker & Docker Compose**
- **Make** (optional, for development commands)

### 1. Clone and Setup

```bash
git clone https://github.com/ianwu0915/SettleChat.git
cd SettleChat

# Copy environment configuration
cp .env.example .env
```

### 2. Start Infrastructure Services

```bash
# Start PostgreSQL and NATS using Docker Compose
docker-compose -f docker/docker-compose.yml up -d

# Verify services are running
docker ps
```

### 3. Configure Environment

```bash
# Edit .env file with your settings
DATABASE_URL=postgres://postgres:secret@localhost:5432/settlechat
NATS_URL=nats://localhost:4222
REDIS_URL=redis://localhost:6379
```

### 4. Run the Application

```bash
# Install dependencies
go mod download

# Run with Make (recommended)
make run

# Or run directly
go run cmd/server/main.go
```

### 5. Access the Application

- **Web Interface**: http://localhost:8080
- **Chat Application**: http://localhost:8080/login.html
- **NATS Monitoring**: http://localhost:8222

## 💻 Development

### Available Commands

```bash
# Code formatting and linting
make format          # Format code with gofmt and goimports
make lint           # Run staticcheck linter
make check          # Run both format and lint

# Testing and benchmarks
go test ./...                           # Run all tests
go test -bench=. ./benchmark           # Run performance benchmarks

# Dependency management
make tidy           # Clean up and verify dependencies
```

### Project Structure

```
settlechat/
├── cmd/server/                 # Application entry point
│   ├── main.go                # Server initialization and routing
│   └── handler/               # HTTP request handlers
│       ├── authHandlers.go    # Authentication endpoints
│       ├── roomHandler.go     # Room management endpoints
│       └── wshandler.go       # WebSocket upgrade handler
├── internal/                  # Core application logic
│   ├── ai/                    # AI integration modules
│   │   ├── agent.go           # AI conversation agent
│   │   ├── manager.go         # AI service management
│   │   ├── command.go         # AI command processing
│   │   ├── summary.go         # Conversation summarization
│   │   └── providers/         # AI provider implementations
│   ├── chat/                  # Real-time chat core
│   │   ├── hub.go            # Connection management hub
│   │   ├── room.go           # Chat room logic
│   │   └── client.go         # WebSocket client handling
│   ├── messaging/             # Event-driven messaging system
│   │   ├── eventbus.go       # Event abstraction layer (wraps NATS)
│   │   └── nats/             # NATS implementation
│   │       ├── nats.go       # Connection management
│   │       ├── publish.go    # Message publishing
│   │       ├── subscribe.go  # Event subscription & routing
│   │       └── nats_topics.go # Topic formatting utilities
│   ├── storage/               # Data persistence
│   │   ├── db.go             # Database connection
│   │   ├── messageStore.go   # Message CRUD operations
│   │   └── user.go           # User management
│   └── event_handlers/        # Event processing
├── web/                       # Frontend assets
│   ├── login.html            # Authentication interface
│   ├── rooms.html            # Room management UI
│   └── chat.html             # Chat interface
├── benchmark/                 # Performance testing
├── docker/                    # Container configuration
└── docs/                      # Documentation
```

## 🤖 AI Features

### Available AI Commands

| Command          | Description                     | Example                                          |
| ---------------- | ------------------------------- | ------------------------------------------------ |
| `/summary`       | Generate conversation summary   | `/summary`                                       |
| `/help`          | Show available commands         | `/help`                                          |
| `/stats`         | Display AI assistant statistics | `/stats`                                         |
| `/clear`         | Clear summary history           | `/clear`                                         |
| `/prompt <text>` | Custom AI processing            | `/prompt analyze the technical issues discussed` |

### AI Integration

SettleChat supports multiple AI providers through a pluggable interface:

```go
type Provider interface {
    GetName() string
    GenerateSummary(ctx context.Context, messages []MessageInput, previousSummary string) (string, error)
    ProcessPrompt(ctx context.Context, messages []MessageInput) (string, error)
}
```

Currently supported providers:

- **DeepSeek**: Cost-effective AI processing
- **Mock Provider**: Development and testing
- **OpenAI**: (Planned) GPT integration

## 📊 Performance & Benchmarks

The application includes comprehensive benchmarking for performance monitoring:

```bash
# Run WebSocket connection benchmarks
go test -bench=BenchmarkWebSocketConnections ./benchmark

# Run concurrent connection tests
go test -bench=BenchmarkConcurrentWebSocketConnections ./benchmark

# Run message throughput benchmarks
go test -bench=BenchmarkMessageThroughput ./benchmark

# Run database performance tests
go test -bench=BenchmarkSaveMessage ./benchmark
```

## 🔧 Configuration

### Environment Variables

```bash
# Database Configuration
DATABASE_URL=postgres://user:password@host:port/database

# NATS Configuration
NATS_URL=nats://localhost:4222

# Redis Configuration (optional)
REDIS_URL=redis://localhost:6379

# Application Environment
ENVIRONMENT=dev|prod
```

### NATS Topic Structure

```
settlechat{env}.{category}.{action}.{roomID}

Examples:
- settlechat.user.joined.room123
- settlechat.message.chat.room123
- settlechat.ai.command.room123
- settlechat.message.history.request.room123
```

## 🐳 Deployment

### Docker Deployment

```bash
# Build application image
docker build -t settlechat .

# Run with Docker Compose
docker-compose -f docker/docker-compose.yml up -d

# Scale the application
docker-compose -f docker/docker-compose.yml up -d --scale app=3
```

### Production Considerations

- **Load Balancing**: Use NGINX or similar for WebSocket load balancing
- **Database**: Configure PostgreSQL with proper connection pooling
- **NATS Clustering**: Set up NATS cluster for high availability
- **Monitoring**: Implement metrics collection and alerting
- **SSL/TLS**: Configure HTTPS and WSS for production

## 🔮 Roadmap

### Upcoming Features

- [ ] **Advanced AI Capabilities**

  - GPT-4 integration
  - Custom AI model training
  - Multi-language support

- [ ] **Enhanced Collaboration**

  - File sharing and media support
  - Video/voice call integration
  - Screen sharing capabilities

- [ ] **Enterprise Features**

  - SSO integration (SAML, OAuth)
  - Advanced permissions and roles
  - Audit logging and compliance

- [ ] **Performance & Scale**

  - Redis caching layer
  - Horizontal scaling improvements
  - WebRTC for peer-to-peer communication

- [ ] **Developer Experience**
  - REST API documentation
  - WebSocket API documentation
  - Client SDKs (Go, Python, JavaScript)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests
4. **Run the test suite**: `go test ./...`
5. **Run benchmarks**: `go test -bench=. ./benchmark`
6. **Commit your changes**: `git commit -m 'Add amazing feature'`
7. **Push to the branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [NATS](https://nats.io) for the excellent messaging system
- [Gorilla WebSocket](https://github.com/gorilla/websocket) for WebSocket implementation
- [PostgreSQL](https://postgresql.org) for reliable data storage
- [DeepSeek](https://deepseek.com) for AI integration capabilities

---

**Built with ❤️ using Go, NATS, and modern web technologies**

For questions, issues, or feature requests, please [open an issue](https://github.com/ianwu0915/SettleChat/issues) or join our community discussions.
