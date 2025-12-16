# GoStack

A production-ready, multi-tenant B2B SaaS starter kit written in Go.  
Designed for scalability, ease of use, and strict tenant isolation.

## Features

- **Multi-Tenancy**: Built-in data isolation using `business_id`.
- **Authentication**: JWT-based auth with Access/Refresh tokens.
- **Advanced RBAC**: Role-Based Access Control with database-backed roles and permissions.
- **Dynamic Menus**: API-driven, permissible menu structure for frontend applications.
- **Worker System**: Lightweight Redis-based background job processing with distributed locking.
- **Observability**: OpenTelemetry (OTel) instrumentation for tracing and metrics.
- **Dependency Injection**: Clean architecture using a robust Container pattern.

## Quick Start

### 1. Clone and Rename
This repo is a template. Use the included script to rename the module to your own path.

```bash
git clone https://github.com/emuthianimbithi/GoStack.git my-saas
cd my-saas
./scripts/rename_module.sh github.com/my-org/my-saas
go mod tidy
```

### 2. Setup Infrastructure
You need **PostgreSQL** and **Redis**.

```bash
# Example via Docker
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15
docker run -d -p 6379:6379 redis:7
```

### 3. Run the Server
The server will auto-migrate the database schema on startup.

```bash
export HTTP_ADDR=:8080
export DB_DSN="host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"

go run cmd/api/main.go
```

### 4. Create Your First Master Admin
Use the seeding script to create the initial system roles and a master admin user.
*Note: Master Admins can impersonate any tenant for support purposes.*

```bash
go run cmd/seed/main.go
```

## Configuration

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `HTTP_ADDR` | Server listen address | `:8080` |
| `DB_DSN` | Postgres connection string | (Localhost defaults) |
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379/0` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel Collector Endpoint | `localhost:4317` |
| `AUTH_ACCESS_TOKEN_DURATION` | Access token TTL | `15m` |
| `AUTH_REFRESH_TOKEN_DURATION` | Refresh token TTL | `168h` (7 days) |

## API Documentation
The API is structured under `/api/v1`.
- **Auth**: `/auth/register`, `/auth/login`
- **Users**: `/api/v1/users`
- **Roles**: `/api/v1/roles`
- **Menus**: `/api/v1/menus`

See `ARCHITECTURE.md` for a deep dive into the code structure.
