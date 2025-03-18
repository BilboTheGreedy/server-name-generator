# Server Name Generator

A Go-based application for generating and managing server names.

## Getting Started

### Prerequisites
- Go 1.21 or later
- PostgreSQL database
- Docker and Docker Compose (optional)

### Setup and Run

1. Clone the repository
```bash
git clone https://github.com/bilbothegreedy/server-name-generator.git
cd server-name-generator
```

2. Start the database (using Docker Compose)
```bash
docker-compose up -d database
```

3. Run the application
```bash
go run cmd/server/main.go
```

4. Access the application at http://localhost:8080

### Environment Variables

See the `.env` file for configuration options.

## Features

- Server name generation and management
- Admin dashboard
- User authentication
- API access
