# Contributing Guide

## Development Workflow

1.  **Fork & Clone**
2.  **Install Dependencies**: `go mod download`
3.  **Run Tests**:
    ```bash
    go test ./...
    ```

## Coding Standards

### 1. No Global State
Do not use global variables for database connections or config. Everything must be passed via the `Container` or constructor injection.

### 2. Service Layer Pattern
- **Handlers** handling HTTP logic (parsing JSON, status codes).
- **Services** handle Business logic (validation, transactions).
- **Repositories** handle SQL.

### 3. Database Changes
- We use **GORM AutoMigrate** for simple iteration.
- For production, consider adding a tool like `golang-migrate` for versioned migrations.

### 4. Adding New Features
1. Define the **Model** in `internal/models`.
2. Create the **Repository** interface & implementation.
3. Create the **Service** method.
4. Create the **Handler** endpoint.
5. Register the route in `internal/server/router`.
