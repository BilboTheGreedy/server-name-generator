# Server Name Generator

## Overview
Server Name Generator is a robust, scalable application for generating and managing server names across different environments.

## Features
- Automated server name generation
- API-driven name reservation and management
- User and API key authentication
- Admin dashboard
- Comprehensive logging and monitoring

## Prerequisites
- Docker
- Docker Compose
- Git

## Installation

### Local Development
1. Clone the repository
```bash
git clone https://github.com/bilbothegreedy/server-name-generator.git
cd server-name-generator
```

2. Environment Configuration
Copy the `.env.example` to `.env` and modify as needed:
```bash
cp .env.example .env
```

3. Start the Application
```bash
docker-compose up --build
```

### Database Migrations
Use golang-migrate for managing database schema:
```bash
# Install migrate CLI
brew install golang-migrate  # macOS
# Or
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create new migration
migrate create -ext sql -dir migrations -seq create_initial_tables

# Run migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/server_names?sslmode=disable" up
```

## API Documentation

### Authentication
- JWT Token Authentication
- API Key Authentication

### Endpoints
- `POST /api/reserve`: Reserve a server name
- `POST /api/commit`: Commit a reservation
- `GET /api/reservations`: List all reservations
- `GET /api/stats`: Get system statistics

### Authorization Scopes
- `read`: View reservations
- `reserve`: Create new reservations
- `commit`: Commit reservations
- `release`: Release committed reservations

## Deployment

### Production Considerations
- Use strong, randomly generated passwords
- Enable HTTPS
- Implement regular backups
- Monitor system health

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database hostname | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `LOG_LEVEL` | Logging verbosity | `info` |

## Backup Strategy
- Daily automated backups
- Backups compressed and stored in `/backups`
- Retention of 7 days of backups

## Health Monitoring
Access the health endpoint at `/api/health` for:
- System status
- Database connectivity
- Runtime metrics

## Logging
Uses structured JSON logging with context and request tracking.

## Contributing
1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Push and create a Pull Request

## License
None

## Support
For issues, please file a GitHub issue