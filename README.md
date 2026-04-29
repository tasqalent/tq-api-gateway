# TASQALENT API Gateway

The **API Gateway** is the single public entry point for TASQALENT. It routes requests to backend microservices, applies cross-cutting middleware (request ID, logging, security headers, CORS), and can enforce JWT authentication/RBAC at the edge.

## What's implemented

### Routing (reverse proxy mounts)

REST mounts (enabled only when the corresponding `*_SERVICE_URL` is set):

- `/auth/*` → `AUTH_SERVICE_URL`
- `/users/*` → `USERS_SERVICE_URL`
- `/gigs/*` → `GIG_SERVICE_URL`
- `/chat/*` → `CHAT_SERVICE_URL`
- `/orders/*` → `ORDER_SERVICE_URL`
- `/reviews/*` → `REVIEW_SERVICE_URL`

WebSocket mounts (enabled only when corresponding `*_WS_URL` is set):

- `/ws/chat/*` → `CHAT_WS_URL`
- `/ws/orders/*` → `ORDER_WS_URL`

Health:

- `GET /healthz` → returns `{"status": "ok"}`

### Middleware (edge behavior)

Applied globally:

- **Request ID**: generates/propagates `X-Request-ID` and forwards it to upstreams.
- **Access logs**: structured logs for method/path/status/duration.
- **Security headers**: basic headers (safe defaults).
- **CORS**: via `rs/cors`, configured by env.

Auth enforcement:

- **Public paths**: configured by `GATEWAY_PUBLIC_PATHS` (default: `/healthz,/auth`).
- **JWT**: enabled when `GATEWAY_JWT_SECRET` is set (HS256 Bearer tokens).
- **RBAC hook**: optional single-role enforcement via `GATEWAY_REQUIRED_ROLE` (expects a `role` claim).

## Configuration

Copy `.env.example` → `.env` and adjust as needed:

```bash
cp .env.example .env
```

Key env vars:

### Core

- `SERVICE_NAME` (default: `tq-api-gateway`)
- `HTTP_ADDR` (default: `:8080`)
- `LOG_LEVEL` (default: `INFO`)
- `GATEWAY_PROXY_TIMEOUT` (default: `30s`)

### Auth / security

- `GATEWAY_JWT_SECRET` (empty disables JWT enforcement)
- `GATEWAY_PUBLIC_PATHS` (CSV of public path prefixes; default: `/healthz,/auth`)
- `GATEWAY_REQUIRED_ROLE` (empty disables RBAC)

### Upstreams (REST)

- `AUTH_SERVICE_URL`
- `USERS_SERVICE_URL`
- `GIG_SERVICE_URL`
- `CHAT_SERVICE_URL`
- `ORDER_SERVICE_URL`
- `REVIEW_SERVICE_URL`

### Upstreams (WebSocket)

- `CHAT_WS_URL`
- `ORDER_WS_URL`
- `GATEWAY_WS_IDLE_TIMEOUT` (default: `5m`)

### CORS

- `CORS_ALLOWED_ORIGINS` (CSV; default: `http://localhost:5173`)
- `CORS_ALLOW_CREDENTIALS` (`true` / `false`)

## Run on host

```bash
go run ./cmd/gateway
```

Verify:

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/healthz/
```

## Run in Docker

Build and run directly:

```bash
docker build -t tq-api-gateway .
docker run --rm -p 8080:8080 --env-file .env tq-api-gateway
```

## Run with local infra (Docker Compose)

```bash
cd ../tq-infra
docker-compose up -d api-gateway
```

### Upstream URL tip (local dev)

- If a backend is running on your **host**, set its URL to `http://host.docker.internal:<port>` so the gateway container can reach it.
- If a backend is running as a **compose service**, set its URL to `http://<service-name>:<port>`.
