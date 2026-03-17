# Example Velo Application

This is an example blog application built with Velo Framework. It demonstrates best practices and common patterns.

## Features

- RESTful API for managing blog posts
- User authentication with JWT
- Database migrations
- Request validation
- Middleware chains
- Service layer pattern

## Project Structure

```
example-app/
├── app/
│   ├── controllers/
│   │   ├── post_controller.go
│   │   └── auth_controller.go
│   ├── models/
│   │   ├── post.go
│   │   └── user.go
│   ├── middleware/
│   │   └── auth.go
│   ├── services/
│   │   ├── post_service.go
│   │   └── user_service.go
│   └── validators/
│       └── post_request.go
├── bin/
│   └── main.go
├── database/
│   ├── migrations/
│   │   ├── 2024_01_01_create_users_table.go
│   │   └── 2024_01_02_create_posts_table.go
│   └── seeders/
├── tests/
├── .env
└── go.mod
```

## Getting Started

1. Copy this directory to a new location
2. Update module name in `go.mod`
3. Install dependencies: `go mod tidy`
4. Configure `.env` file
5. Run migrations: `velo migrate run`
6. Start server: `go run bin/main.go`

## API Endpoints

### Posts

- `GET /posts` - List all posts
- `POST /posts` - Create a new post
- `GET /posts/:id` - Get a specific post
- `PUT /posts/:id` - Update a post
- `DELETE /posts/:id` - Delete a post

### Authentication

- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login user
- `POST /auth/logout` - Logout user

## Testing

Run the test suite:

```bash
go test ./...
```

## Database

This example uses SQLite for simplicity. To use PostgreSQL:

1. Update `.env`:

```env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_DATABASE=blog_app
```

2. Update `bin/main.go` to use PostgreSQL driver

## Learn More

- [Framework Documentation](../README.md)
- [Getting Started Guide](../GETTING_STARTED.md)
- [Architecture Guide](../ARCHITECTURE.md)
