# Car Rental Platform

A full-stack car rental management system built with Go microservices and Next.js.

## Architecture

The platform follows a microservice architecture with four backend services, an API gateway, and a frontend application. Services communicate via REST HTTP, gRPC, and NATS messaging.

```
Browser -> Traefik/Nginx -> Frontend (Next.js :3000)
                         -> API Gateway (Go :8080)
                               -> Identity Service (Go :8080)
                               -> Vehicle Service  (Go :8081)
                               -> Booking Service  (Go :8082)
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| api-gateway | 8080 | Reverse proxy — routes requests to backend services, handles CORS |
| identity-and-notification-service | 8080, 50051 | User registration, login, JWT auth, profile management |
| vehicle-inventory-service | 8081, 50052 | Vehicle CRUD, locations, maintenance records |
| booking-and-pricing-service | 8082, 50053 | Booking lifecycle, pricing tiers, availability checking |
| frontend | 3000 | Next.js application with Tailwind CSS |

### Data Flow

1. User logs in via Identity Service — receives JWT token
2. Token is stored in browser localStorage and sent with each request
3. API Gateway proxies requests to the correct backend service
4. Booking Service calls Vehicle Service via gRPC for availability checks
5. NATS is used for async event messaging between services

## Tech Stack

**Backend (Go)**
- Gin — HTTP router
- gRPC — internal service-to-service communication
- pgx/v5 — PostgreSQL driver (no ORM)
- golang-jwt/v5 — JWT authentication
- go-redis/v9 — Redis caching
- nats.go — async messaging via NATS
- prometheus/client_golang — metrics collection

**Frontend (TypeScript)**
- Next.js 15 — App Router, server components
- Tailwind CSS v3 — utility-first styling
- lucide-react — icons
- Cloudflare — DNS and TLS termination

**Infrastructure**
- Docker + Docker Compose — containerization
- PostgreSQL 17 — primary database
- Redis 8 — caching layer
- NATS 2.10 — message broker
- Traefik — reverse proxy for production
- Prometheus — metrics scraping
- Grafana — monitoring dashboard

## Getting Started (Development)

Prerequisites: Docker, Docker Compose.

```bash
# Clone the repository
git clone <repo-url>
cd car-rental-platform

# Start everything
docker compose up --build -d

# Verify
curl http://localhost:80/health
curl http://localhost:80/api/v1/vehicles
```

Open http://localhost:3001 in your browser.

### Default Credentials

- PostgreSQL: `postgres` / `postgres`
- Grafana: `admin` / `admin` (http://localhost:3000)

## Production Deployment

This project is deployed at https://vehicle-nurashi.abzy.kz using Traefik as the reverse proxy.

### Prerequisites on Server

- Docker and Docker Compose
- Traefik running with `proxy-net` network
- PostgreSQL and Redis running externally (separate containers)

### Deploy

```bash
# Copy project to server
scp -r ./final user@server:~/final

# On the server
cd ~/final
docker compose -f docker-compose.prod.yml up --build -d
```

### Environment Configuration

Each service has a `.env.example` file with placeholder values. Copy and customize:

```bash
cp identity-and-notification-service/.env.example identity-and-notification-service/.env
# Edit .env with real database credentials
```

Traefik routes are configured via Docker labels in `docker-compose.prod.yml`. The API gateway handles `/api` paths and the frontend handles everything else, both on the same domain.

## API Overview

Base URL: `http://localhost:80/api/v1` (dev) or `https://vehicle-nurashi.abzy.kz/api/v1` (prod)

### Identity

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /identity/register | No | Register new user |
| POST | /identity/login | No | Login, receive JWT |
| POST | /identity/verify-email | No | Verify email address |
| POST | /identity/resend-verification | No | Resend verification email |
| GET | /identity/profile/:id | JWT | Get user profile |
| PUT | /identity/profile/:id | JWT | Update user profile |

### Vehicles

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /vehicles | No | List vehicles (paginated) |
| GET | /vehicles/:id | No | Get vehicle details |
| POST | /vehicles | JWT | Create vehicle |
| PUT | /vehicles/:id | JWT | Update vehicle |
| DELETE | /vehicles/:id | JWT | Delete vehicle |
| PUT | /vehicles/:id/status | JWT | Update vehicle status |
| GET | /locations | No | List rental locations |
| POST | /locations | JWT | Create location |
| GET | /vehicles/:id/maintenance | No | Get maintenance records |
| POST | /vehicles/:id/maintenance | JWT | Create maintenance record |

### Bookings

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /bookings/calculate-price | No | Calculate rental price |
| GET | /bookings/availability | No | Check vehicle availability |
| POST | /bookings | JWT | Create a booking |
| GET | /bookings/:id | JWT | Get booking details |
| DELETE | /bookings/:id | JWT | Cancel booking |
| GET | /users/:user_id/bookings | JWT | List user bookings |

## Project Structure

```
api-gateway/                  # Reverse proxy (Gin)
  cmd/main.go                 # Entry point
  internal/config/            # Environment config
  internal/health/            # Health check endpoints
  internal/middleware/        # CORS
  internal/proxy/             # Proxy handler, route setup

booking-and-pricing-service/  # Booking and pricing (Go)
  cmd/main.go                 # Entry point
  internal/api/               # HTTP + gRPC handlers
  internal/domain/            # Domain models
  internal/config/            # Environment config
  internal/db/                # Database connection + migrations
  internal/messaging/         # NATS pub/sub
  internal/repository/        # Database queries (pgx)
  internal/service/           # Business logic

identity-and-notification-service/  # Auth and profiles (Go)
  (same layered structure as above)

vehicle-inventory-service/    # Vehicle management (Go)
  (same layered structure as above)

frontend/                     # Web UI (Next.js)
  src/app/                    # App Router pages
  src/app/vehicles/           # Vehicle listing and detail
  src/app/bookings/           # User bookings
  src/app/login/              # Login page
  src/app/register/           # Registration page
  src/app/profile/            # User profile
  src/components/             # Shared components
  src/lib/api.ts              # API client

docker-compose.yml            # Local development setup
docker-compose.prod.yml       # Production deployment
prometheus.yml                # Prometheus scrape config
```

## Database Schema

Each microservice has its own database with migrations applied automatically on startup.

- **identity** — users, driver_licenses, email_verifications, notifications
- **vehicle** — vehicles, locations, maintenance_records
- **booking** — bookings, pricing_tiers, refunds

Tables use `VARCHAR(36)` for primary keys (UUIDs) and `TIMESTAMPTZ` for timestamps.

## Known Limitations

- Auth middleware does not strip the `Bearer ` prefix from the Authorization header. The frontend sends the raw JWT token as a workaround.
- `created_at` and `updated_at` timestamps return Go zero values instead of database defaults on INSERT responses.
- Several booking endpoints (payment, extend, refund) are registered in the gateway but not yet implemented in the booking service HTTP handler.
- The frontend JWT is stored in `localStorage` with no HTTP-only cookie option.
