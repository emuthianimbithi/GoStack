# Architecture Guide

This document explains the design decisions behind the SaaS Core Framework.

## 1. Directory Structure

We follow the Standard Go Project Layout:

- `cmd/api`: The entrypoint. Wires everything together.
- `internal/`: Private application code.
  - `container`: **Dependency Injection**. The Brain of the app. Initializes all services.
  - `server/`: HTTP layer.
    - `handlers`: Controllers. Parse Request -> Call Service -> Return Response.
    - `services`: Business Logic.
    - `repositories`: Database Access.
    - `middleware`: Cross-cutting concerns (Auth, Tenancy, Logging).
  - `models`: Domain entities (GORM structs).
  - `permissions`: RBAC logic.
  - `worker`: Background job processor.

## 2. Multi-Tenancy Strategy

We use a **Shared Database, Separate Schema** approach (logically separated by `business_id`).

### The `Business` Entity
- Every tenant is a `Business`.
- Every `User` belongs to one `Business`.
- Users have a **Scoped Role** (`role_id` + `business_id`) within that business.

### Isolation
Isolation is enforced at the **HTTP Middleware** level:
1. `TenantMiddleware` extracts `BusinessID` from the JWT.
2. It sets it in the `context.Context`.
3. Handlers/Services must pass this `business_id` to Repositories to filter queries.

**Impersonation**: `MasterAdmin` users can override the context `BusinessID` using the `X-Impersonate-Business-ID` header, allowing them to view the app *as* a specific tenant.

## 3. Authentication & RBAC

### JWT Claims
Tokens contain:
- `sub`: User ID
- `bid`: Business ID
- `role`: Role Name (e.g., "admin")

### Permissions
We do not hardcode Role checks (e.g. `if role == "admin"`). Instead, we verify **Permissions**.
- **Roles** have many **Permissions** (linked to `APIResources`).
- Middleware checks: "Does User's Role have permission for `GET /api/v1/users`?"
- This allows creating custom roles on the fly without changing code.

## 4. Why Dependency Injection?
We use a `Container` struct in `internal/container` to explicitly manage dependencies.
- **Pros**: Easy testing (mocking interfaces), clear initialization order, no global state.
- **Cons**: More boilerplate than `init()` functions.
- **Verdict**: Essential for long-term maintainability of a SaaS codebase.

## 5. Billing & Ledger
We use a **Hybrid Adapter Pattern** for billing:
- **LedgerEntry**: The Source of Truth. An immutable log of every financial transaction (Credit/Debit) stored in your DB.
- **Subscriptions**: Synced state from providers (Stripe/M-Pesa) but tracked locally for fast access checks.
- **Provider Adapters**: The `BillingService` switches logic based on the provider string, keeping payment complexity isolated.

## 6. Communication
- **EmailService**: Wrapper around SendGrid. Supports tenant-specific "From" names/addresses if configured in `BusinessSettings`.
- **NotificationService**: Central hub for user alerts. Stores in DB first (polled by UI), with hooks for Push Notifications (FCM).

## 7. Webhooks (Outbound)
Tenants can subscribe to events (e.g. `order.created`) via `POST /webhooks`.
- **System**: When an event occurs, `WebhookService.DispatchEvent` is called.
- **Security**: payloads are signed with `HMAC-SHA256` using a secret unique to that endpoint.
- **Reliability**: Failures are logged in `WebhookDelivery` for audit/debug purposes.
