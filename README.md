# Go Microservice Framework

A comprehensive Go framework for building production-ready microservices with built-in support for web servers, databases, caching, message queues, WebSockets, and more.

## Features

- 🚀 **Web Server**: HTTP router with middleware support (using httprouter)
- 🗄️ **Database**: PostgreSQL and SQLite support with GORM
- ⚡ **Cache**: Redis integration for caching
- 📨 **Message Bus**: AWS SQS integration for async messaging
- 🔌 **WebSockets**: Real-time bidirectional communication
- 🔐 **Authentication**: JWT and Google OAuth support
- 📝 **Logging**: Structured logging with Zap
- 🔔 **Alerts**: Slack and SMS notifications
- 🌐 **Proxy Support**: HTTP and SOCKS5 proxy handlers
- 🕷️ **Web Scraping**: Built-in scraper utilities with Colly and GoQuery

## Installation

```bash
go get github.com/unluckythoughts/go-microservice
```

## Quick Start

```go
package main

import (
    "github.com/unluckythoughts/go-microservice"
    "github.com/unluckythoughts/go-microservice/tools/web"
)

func handler(r web.Request) (interface{}, error) {
    return map[string]string{"message": "Hello, World!"}, nil
}

func main() {
    opts := microservice.Options{
        Name:        "my-service",
        EnableDB:    true,
        EnableCache: true,
    }
    
    s := microservice.New(opts)
    s.HttpRouter().GET("/hello", handler)
    s.Start()
}
```

## Configuration

Configure the service using environment variables:

### Service Configuration
- `SERVICE_NAME`: Service name (default: "true")
- `SERVICE_ENABLE_DB`: Enable database (default: true)
- `SERVICE_DB_TYPE`: Database type - "postgresql" or "sqlite" (default: "postgresql")
- `SERVICE_ENABLE_CACHE`: Enable Redis cache (default: false)
- `SERVICE_ENABLE_BUS`: Enable message bus (default: false)

### Database Configuration
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_HOST`: Database host (default: "localhost")
- `DB_PORT`: Database port (default: "5432")

### Web Server Configuration
- `WEB_PORT`: HTTP server port (default: "8080")
- `WEB_CORS`: Enable CORS (default: false)

### Cache Configuration
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port
- `REDIS_PASSWORD`: Redis password

### Message Bus Configuration
- `SQS_REGION`: AWS region
- `SQS_ACCESS_KEY_ID`: AWS access key
- `SQS_SECRET_ACCESS_KEY`: AWS secret key

## Available Tools

### Web (`tools/web`)
- HTTP router with middleware support
- Request/response handling
- JWT authentication
- Session management
- WebSocket integration
- Custom validators
- Proxy handlers (HTTP and SOCKS5)

### Database (`tools/psql`, `tools/sqlite`)
- PostgreSQL support with GORM
- SQLite support with GORM
- Database utilities

### Cache (`tools/cache`)
- Redis client wrapper
- Caching utilities

### Message Bus (`tools/bus`)
- AWS SQS integration
- Message publishing and consumption

### Authentication (`tools/auth`)
- JWT token generation and validation
- Google OAuth integration
- User authentication middleware
- Session management

### WebSockets (`tools/sockets`)
- WebSocket server
- Connection management
- Custom handlers
- Worker pools

### Logging (`tools/logger`)
- Structured logging with Zap
- Context-aware logging

### Alerts (`utils/alerts`)
- Slack notifications
- SMS/Text alerts

### Utilities (`utils`)
- Encryption utilities
- Database helpers
- Environment variable loading
- Web scraping tools

## API Structure

### Service Interface

```go
type IService interface {
    Start()                                                     // Start the service
    HttpRouter() web.Router                                     // Get HTTP router
    SocketRegister(method string, handler sockets.Handler)     // Register WebSocket handler
    GetDB() *gorm.DB                                           // Get database instance
    GetCache() *redis.Client                                   // Get cache instance
    GetBus() bus.IBus                                          // Get message bus instance
    GetAlerts() (*alerts.SlackClient, *alerts.TextClient)      // Get alert clients
    GetLogger() *zap.Logger                                    // Get logger instance
}
```

### HTTP Router

```go
// Define routes with optional middleware
router.GET("/path", middleware1, middleware2, handler)
router.POST("/path", handler)
router.PUT("/path", handler)
router.DELETE("/path", handler)
router.PATCH("/path", handler)
```

### Middleware

```go
func myMiddleware(r web.MiddlewareRequest) error {
    // Access logger
    r.GetContext().Logger().Info("middleware executing")
    
    // Set context values
    r.SetContextValue("key", "value")
    
    return nil
}
```

### Handlers

```go
func myHandler(r web.Request) (interface{}, error) {
    // Get context values
    val := r.GetContext().Value("key")
    
    // Access logger
    r.GetContext().Logger().Info("handler executing")
    
    // Return response (will be JSON encoded)
    return map[string]interface{}{"status": "success"}, nil
}
```

## Examples

See the [examples](examples/) directory for complete examples:
- [Microservice Example](examples/microservice/main.go): Basic microservice setup
- [Scraper Example](examples/scrapper/main.go): Web scraping utilities

## Dependencies

Key dependencies include:
- `gorm.io/gorm`: ORM for database operations
- `github.com/go-redis/redis/v8`: Redis client
- `github.com/aws/aws-sdk-go`: AWS SDK for SQS
- `go.uber.org/zap`: Structured logging
- `github.com/julienschmidt/httprouter`: HTTP router
- `github.com/golang-jwt/jwt/v5`: JWT authentication
- `github.com/gobwas/ws`: WebSocket implementation
- `github.com/gocolly/colly/v2`: Web scraping
- `golang.org/x/oauth2`: OAuth2 client

## Project Structure

```
.
├── service.go              # Main service implementation
├── tools/                  # Core tools and utilities
│   ├── auth/              # Authentication and authorization
│   ├── bus/               # Message bus (SQS)
│   ├── cache/             # Redis cache
│   ├── logger/            # Logging utilities
│   ├── psql/              # PostgreSQL database
│   ├── sockets/           # WebSocket support
│   ├── sqlite/            # SQLite database
│   └── web/               # Web server and routing
├── utils/                  # Utility functions
│   ├── alerts/            # Notification utilities
│   └── scrapper/          # Web scraping tools
└── examples/              # Example implementations
```

## TODO

- [ ] Refactor codebase
- [ ] Add Swagger/OpenAPI specification
- [ ] Add comprehensive unit tests
- [ ] Add more examples and documentation
- [ ] Add metrics and monitoring support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.