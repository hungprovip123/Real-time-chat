# Week 3 Chat Application

Real-time chat application built with Go, WebSocket, Redis, and Bootstrap.

## Features

- **Real-time messaging** với WebSocket
- **Multiple chat rooms** (General, Tech, Random)
- **Message history** lưu trong Redis
- **Online users tracking**
- **Rate limiting** chống spam (10 requests/60 seconds)
- **Username validation** (không trùng trong room)
- **Auto-reconnect** khi mất kết nối

## Tech Stack

- **Backend**: Go, Gin framework, Gorilla WebSocket
- **Database**: Redis (message storage, rate limiting, online users)
- **Frontend**: HTML, Bootstrap, JavaScript
- **Architecture**: Clean Architecture với Goroutines & Channels

## Quick Start

1. **Clone repository:**
```bash
git clone <repository-url>
cd week3
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Start Redis server** (required):
```bash
# Install Redis and run
redis-server
```

4. **Run application:**
```bash
go run main.go
```

5. **Open browser:**
```
http://localhost:8080
```

## Environment Variables

Application sử dụng default values, có thể override:

```bash
PORT=8080
REDIS_URL=redis://localhost:6379
JWT_SECRET=week3-secret-key
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_WINDOW=60
```

## API Endpoints

- `GET /` - Chat interface
- `GET /api/v1/ws` - WebSocket connection
- `GET /api/v1/rooms/{roomId}/messages` - Get message history
- `GET /api/v1/rooms/{roomId}/users` - Get online users
- `POST /api/v1/messages` - Send message (fallback)

## Project Structure

```
week3/
├── main.go              # Entry point
├── config/              # Configuration
├── handlers/            # HTTP handlers
├── models/              # Data structures
├── redis/               # Redis client
├── websocket/           # WebSocket logic
└── templates/           # HTML templates
```

## How it Works

1. **Client connects** via WebSocket với username & room
2. **Hub manages** all connections và broadcasting
3. **Messages persist** trong Redis
4. **Real-time updates** qua WebSocket channels
5. **Goroutines handle** concurrent connections

Built for Week 3 assignment - focuses on Go concurrency patterns và real-time communication. 