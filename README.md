# NuvelaOne API

A multi-tenant booking management system for small businesses. Staff can manage bookings while customers can book appointments through a public booking page.

## Features

- **Multi-tenant architecture** - Each business has its own isolated data with a custom slug (e.g., `/salon-slug`)
- **Staff management** - Owners can invite team members with different roles (owner, admin, member)
- **Service management** - Define services with duration, buffer time, and cost
- **Booking system** - Staff create, update, and manage bookings; customers book via public page
- **Working hours** - Configurable per-tenant weekly schedule

## Stack

| Concern        | Library                  |
|----------------|--------------------------|
| Router         | chi                      |
| Database       | PostgreSQL + pgx/v5      |
| Auth           | JWT (golang-jwt)         |
| Config         | godotenv                 |
| Logging        | zap                      |
| Migrations     | goose                    |

## Project Structure

```
.
├── cmd/api/          # Application entry point
├── internal/
│   ├── domain/       # Entities and DTOs
│   ├── handler/     # HTTP handlers
│   ├── service/     # Business logic
│   ├── repository/  # Database layer
│   └── middleware/  # Auth, logging, recovery
├── migrations/       # SQL migrations
├── config/          # Configuration
└── pkg/             # Shared utilities
```

## Getting Started

```bash
# Install dependencies
go mod tidy

# Copy and configure environment
cp .env.example .env

# Run database migrations
make migrate-up

# Start development server
make run
```

## Environment Variables

| Variable      | Description                    |
|---------------|--------------------------------|
| DATABASE_URL | PostgreSQL connection string  |
| JWT_SECRET   | Secret for JWT signing         |
| PORT         | Server port (default 8080)    |

## Commands

| Command          | Description                    |
|------------------|--------------------------------|
| `make run`       | Start development server      |
| `make build`     | Build binary to `bin/api`     |
| `make test`      | Run tests with race detector  |
| `make lint`      | Run golangci-lint             |
| `make migrate-up` | Apply migrations           |
| `make migrate-down` | Rollback migrations       |
| `make docker-up` | Start containers            |

## Database Schema

- **tenants** - Businesses/organizations with unique slugs
- **users** - Staff members belonging to a tenant (role: owner, admin, member)
- **customers** - Clients of a tenant
- **services** - Bookable services (duration, cost, visibility)
- **user_services** - Which staff can perform which services
- **events** - Bookings with status (pending, confirmed, cancelled, completed)
- **working_hours** - Tenant's weekly schedule
- **user_invitations** - Pending team invitations
- **addresses** - Tenant addresses

