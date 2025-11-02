# Example Microservice

This example demonstrates how to use the `go-microservice` framework to build a complete web application with REST API endpoints, socket.io real-time communication, and database operations.

## Features

This example implements a user management system with the following features:

- **User Authentication**: Login and signup with JWT tokens
- **User Management**: Get, update, delete user accounts
- **User Settings**: Customizable user preferences
- **Real-time Communication**: Socket.io for live updates and room-based messaging
- **Database Integration**: PostgreSQL with GORM

## Architecture

The example follows a clean architecture pattern inspired by the table-app backend:

```
example/
├── main.go              # Application entry point
├── go.mod              # Go module definition
├── models/             # Data models and structures
│   └── models.go       # User, UserSettings models
├── db/                 # Database layer
│   └── db.go          # Database operations (CRUD)
├── service/            # Business logic layer
│   └── service.go     # Service methods and handlers
└── api/               # API layer
    ├── routes.go      # HTTP route registration
    └── handlers.go    # WebSocket event handlers
```

## API Endpoints

### Authentication
- `POST /api/v1/login` - User login
- `POST /api/v1/signup` - User registration  
- `POST /api/v1/logout` - User logout (requires auth)

### User Management
- `GET /api/v1/user` - Get current user info (requires auth)
- `PUT /api/v1/user` - Update user profile (requires auth)
- `DELETE /api/v1/user` - Delete user account (requires auth)

### User Settings
- `GET /api/v1/user/settings` - Get user settings (requires auth)
- `PUT /api/v1/user/settings` - Update user settings (requires auth)

### Health Check
- `GET /api/v1/health` - Service health status

## WebSocket Events

### Connection Management
- `connection` - User connects to socket
- `disconnect` - User disconnects from socket
- `welcome` - Welcome message sent on connection

### Real-time Features
- `ping/pong` - Basic connectivity test
- `join_room` - Join a named room for group communication
- `leave_room` - Leave a specific room
- `room_message` - Send message to all users in a room
- `broadcast` - Send message to all connected users

## Data Models

### User
```go
type User struct {
    ID           uint      `json:"id"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    Password     string    `json:"password,omitempty"`
    MobileNumber string    `json:"mobile_number"`
    Role         UserRole  `json:"role"`
    IsVerified   bool      `json:"is_verified"`
    LastLogin    time.Time `json:"last_login"`
}
```

### Item
```go
type Item struct {
    ID          uint   `json:"id"`
    UserID      uint   `json:"user_id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Price       uint   `json:"price"`
    Category    string `json:"category"`
    IsActive    bool   `json:"is_active"`
}
```

### UserSettings
```go
type UserSettings struct {
    ID                   uint   `json:"id"`
    UserID               uint   `json:"user_id"`
    Theme                string `json:"theme"`
    Language             string `json:"language"`
    NotificationsEnabled bool   `json:"notifications_enabled"`
    TimeZone             string `json:"time_zone"`
}
```

## Environment Variables

Set these environment variables before running:

```bash
export SESSION_SECRET="your-jwt-secret-key"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="password"
export DB_NAME="example_db"
```

## Running the Example

1. **Set up PostgreSQL database**
   ```bash
   createdb example_db
   ```

2. **Set environment variables**
   ```bash
   export SESSION_SECRET="your-secret-key-here"
   # ... other DB variables
   ```

3. **Run the application**
   ```bash
   cd example
   go mod tidy
   go run main.go
   ```

4. **Test the API**
   ```bash
   # Health check
   curl http://localhost:8080/api/v1/health
   
   # Register user
   curl -X POST http://localhost:8080/api/v1/signup \
     -H "Content-Type: application/json" \
     -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'
   
   # Login
   curl -X POST http://localhost:8080/api/v1/login \
     -H "Content-Type: application/json" \
     -d '{"email":"john@example.com","password":"password123"}'
   ```

## WebSocket Testing

Connect to `ws://localhost:8080/socket.io/` and send:

```javascript
// Join a room
socket.emit('join_room', { room: 'general' });

// Send message to room
socket.emit('room_message', { 
  room: 'general', 
  message: 'Hello everyone!' 
});

// Broadcast to all users
socket.emit('broadcast', { 
  message: 'Global announcement' 
});
```

## Key Features Demonstrated

1. **Clean Architecture**: Separation of concerns with distinct layers
2. **JWT Authentication**: Secure token-based authentication
3. **Middleware**: Authorization middleware for protected routes
4. **Database Operations**: Full CRUD with GORM
5. **Real-time Communication**: Socket.io integration
6. **Error Handling**: Comprehensive error responses
7. **Input Validation**: Request validation and sanitization
8. **Password Security**: Bcrypt password hashing
9. **User Ownership**: Resource access control
10. **Default Settings**: Automatic user settings creation

## Extending the Example

You can extend this example by adding:

- **File Upload**: Image/document upload for items
- **Search & Filtering**: Advanced item search capabilities
- **Pagination**: Large dataset handling
- **Caching**: Redis integration for performance
- **Rate Limiting**: API rate limiting middleware
- **Email Verification**: User email verification flow
- **OAuth Integration**: Google/Apple login
- **Audit Logging**: Track user actions
- **Notifications**: Push notifications system
- **Multi-tenancy**: Support for organizations/teams

This example provides a solid foundation for building scalable microservices with the go-microservice framework.