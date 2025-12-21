# DeliveryGO

- **A production minded Go based multi package batching delivery orchestration platform** combining routing (OSRM), event-driven matching (Kafka), fast state (Redis), and persistent storage (Postgres).

---

<!-- Tech stack badges -->

[![Go](https://img.shields.io/badge/Go-1.23-blue?logo=go&logoColor=white)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-%232496ED.svg?logo=docker&logoColor=white)](https://www.docker.com)
[![Postgres](https://img.shields.io/badge/Postgres-15-blue?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-orange?logo=redis&logoColor=white)](https://redis.io)
[![Kafka](https://img.shields.io/badge/Kafka-7.5.0-black?logo=apachekafka&logoColor=white)](https://kafka.apache.org)
[![OSRM](https://img.shields.io/badge/OSRM-routing-success?logo=mapbox&logoColor=white)](https://github.com/Project-OSRM/osrm-backend)

---

## Overview

DeliveryGO is an event-driven microservice style codebase written in Go that demonstrates how to coordinate orders, batches, and riders with:

- Routing via OSRM 
- Persistent storage in PostgreSQL (migrations in `migrations/`)
- Asynchronous messaging with Kafka
- Fast ephemeral state in Redis
- Modular app entrypoints in `cmd/` (API, worker, simulator)

Key directories:

- `cmd/api` — HTTP API server
- `cmd/worker` — background worker (matching, batch processing)
- `cmd/simulator/riders` — small simulator to generate rider events
- `internal/` — application logic, domain models, stores, transports
- `data/` — contains OSRM files used by the `osrm` container
- `migrations/` — SQL migrations applied by Flyway in the compose setup
- `api/openapi.yaml` — API schema

---
# Architecture

## Overview

```mermaid
flowchart LR
    Client[Client App] --> API[HTTP API]
    API --> Storage[(Storage Layer)]
    API --> Events[[Event Bus]]
    Events --> Workers[Background Workers]
    Workers --> Routing[OSRM Engine]
```

---

## 1. Order Creation & Assignment Flow

```mermaid
flowchart TD
    Client[Client / Frontend] -->|POST /orders| API[HTTP API]
    
    API -->|1. Create Order| PG_Orders[(Postgres: orders)]
    API -->|2. Convert to H3| H3[H3 Cell]
    API -->|3. Lookup paths| Redis_H3[(Redis: h3 → path_ids)]
    
    Redis_H3 -->|Paths exist| Match[Path Matching Service]
    Redis_H3 -->|No paths| Kafka_Provision[[Kafka: path.provision]]
    
    Match -->|4. Select best path| API
    API -->|5. Assign to batch| PG_Batches[(Postgres: batches)]
    API -->|6. Emit event| Kafka_Assign[[Kafka: batches.assigned]]
```

**Key Components:**
- **H3 Indexing**: Convert lat/lon to H3 cells for efficient spatial lookup
- **Hot Path Cache**: Redis stores pre-computed paths for each H3 cell
- **Path Matching**: Selects optimal path based on distance and capacity

---

## 2. Path Provisioning & Optimization

```mermaid
flowchart TD
    Kafka_Provision[[Kafka: path.provision]] -->|consume| PathWorker[Path Provision Worker]
    
    PathWorker -->|Request route| OSRM[OSRM Engine]
    OSRM -->|Return optimized route| PathWorker
    
    PathWorker -->|Store template| PG_Paths[(Postgres: path_templates)]
    PathWorker -->|Index by H3| Redis_H3[(Redis: h3 → path_ids)]
    
    Kafka_Refine[[Kafka: routes.refine]] -->|consume| OptimizerWorker[Optimizer Worker]
    
    OptimizerWorker -->|Persist mapping| PG_BatchPath[(Postgres: batch_paths)]
    OptimizerWorker -->|Bind batches| Redis_Path[(Redis: path → batches)]
    OptimizerWorker -->|Emit refined| Kafka_Refined[[Kafka: routes.refined]]
    
    Kafka_Refined -->|consume| RegretWorker[Regret Worker]
    RegretWorker -->|Log metrics| Metrics[(Metrics/Logs)]
```

**Key Components:**
- **OSRM**: Open Source Routing Machine for route optimization
- **Path Templates**: Pre-computed routes stored for reuse
- **Regret Analysis**: Measures quality of path assignments over time

---

## 3. Rider Assignment & Tracking

```mermaid
flowchart TD
    Kafka_Refined[[Kafka: routes.refined]] -->|consume| RiderWorker[Rider Assignment Worker]
    
    RiderWorker -->|1. Find nearest| Redis_Riders[(Redis GEO: riders:available)]
    RiderWorker -->|2. Assign rider| PG_Riders[(Postgres: riders + batch_riders)]
    RiderWorker -->|3. Remove from pool| Redis_Riders
    
    RiderSim[Rider Simulator] -->|POST /riders/:id/location| API[HTTP API]
    API -->|Update location| Redis_Riders
    
    API -->|Broadcast updates| WS[WebSocket Hub]
    WS -->|Real-time location| Client[Client App]
```

**Key Components:**
- **Redis GEO**: Efficient geospatial queries for nearest rider lookup
- **WebSocket Hub**: Real-time location updates to clients
- **Rider Simulator**: Testing tool that simulates GPS movements

---

## 4. Delivery Confirmation Flow

```mermaid
flowchart TD
    RiderSim[Rider / Simulator] -->|POST /batches/:id/confirm-delivery| API[HTTP API]
    
    API -->|1. Get batch path| PG_BatchPath[(Postgres: batch_paths)]
    API -->|2. Get drop location| Location[Drop Coordinates]
    
    API -->|3. Verify distance| Haversine[Haversine Check]
    Haversine -->|< 30m| Valid[Valid Delivery]
    Haversine -->|> 30m| Invalid[409 Conflict]
    
    Valid -->|4. Update status| PG_Riders[(Postgres: batch_riders)]
    Valid -->|5. Mark delivered| PG_Batches[(Postgres: batches)]
    Valid -->|6. Notify client| WS[WebSocket]
```

**Key Components:**
- **Haversine Distance**: Validates rider is within 30m of drop location
- **Transaction**: Ensures atomic delivery confirmation across tables
- **Real-time Notification**: Client receives immediate delivery confirmation

---

## 5. Data Architecture

```mermaid
flowchart TD
    subgraph Postgres
        PG_Orders[(orders)]
        PG_Batches[(batches)]
        PG_BatchOrders[(batch_orders)]
        PG_Paths[(path_templates)]
        PG_BatchPath[(batch_paths)]
        PG_Riders[(riders)]
        PG_BatchRiders[(batch_riders)]
    end
    
    subgraph Redis
        Redis_H3[(h3 → path_ids)]
        Redis_Path[(path → batches)]
        Redis_Riders[(GEO: riders:available)]
    end
    
    subgraph Kafka
        Kafka_Provision[[path.provision]]
        Kafka_Assign[[batches.assigned]]
        Kafka_Refine[[routes.refine]]
        Kafka_Refined[[routes.refined]]
    end
```

**Storage Strategy:**
- **Postgres**: Source of truth for all entities and relationships
- **Redis**: Hot path for geospatial queries and real-time state
- **Kafka**: Event bus for async processing and worker coordination

---

## 6. Technology Stack

| Layer | Technology |
|-------|-----------|
| **API** | Go (Chi router) |
| **Database** | PostgreSQL with PostGIS |
| **Cache** | Redis with GEO commands |
| **Message Queue** | Kafka |
| **Routing Engine** | OSRM |
| **Geospatial** | H3 (Uber's hexagonal indexing) |
| **Real-time** | WebSockets |

---

## Key Design Decisions

### 1. Hot Path Architecture
Pre-compute and cache common routes to minimize latency for order assignment.

### 2. H3 Geospatial Indexing
Use Uber's H3 to create uniform spatial regions for efficient path lookup.

### 3. Event-Driven Workers
Decouple heavy computation (routing, optimization) from API response time.

### 4. Redis GEO for Riders
Leverage Redis's built-in geospatial commands for O(log N) nearest rider queries.

### 5. Regret-Based Learning
Continuously measure and improve path assignment quality over time.

---

## Getting Started

See the main [README.md](../README.md) for setup instructions and API documentation.

## Quickstart (local)

Prereqs:

- Docker & Docker Compose
- Go 1.23+ (toolchain indicates 1.24.x is used in the repo)

1) Start infra (Postgres, Kafka, Redis, OSRM)

```bash
docker compose up -d
```

2) Set environment variables (example `.env`)

```env
POSTGRES_DSN=postgres://deliverygo_user:delivery@localhost:5433/deliverygo?sslmode=disable
REDIS_ADDR=localhost:6380
OSRM_ADDR=http://localhost:5000
KAFKA_BROKERS=localhost:9092
HTTP_ADDR=:8000
```

3) Run the API

```bash
go run ./cmd/api
```

4) Run a worker (in another terminal)

```bash
go run ./cmd/worker
```

5) (Optional) Run the rider simulator

```bash
go run ./cmd/simulator/riders
```

Notes:

- The compose file exposes Postgres on host port `5433` and Redis on `6380` — match these in `POSTGRES_DSN` and `REDIS_ADDR`.
- Flyway migrations are included as a `db_migrations` service in `docker-compose.yml` and will run against the `postgres` container.

---

## Environment

The app reads env vars in `internal/app/config.go`. Minimal required variables:

- `POSTGRES_DSN` — Postgres connection string (required)
- `OSRM_ADDR` — HTTP address of OSRM service (required)
- `REDIS_ADDR` — Redis address (default `localhost:6379`)
- `KAFKA_BROKERS` — comma-separated brokers (default `localhost:9092`)
- `HTTP_ADDR` — address for the API server (default `:8000`)

---

## API

API schema is available at [api/openapi.yaml](api/openapi.yaml).

---

## Development tips

- Use `go run ./cmd/...` to run individual entrypoints during development.
- Database migrations are in `migrations/` and applied by the compose `db_migrations` service.
- OSRM expects the prepared `.osrm` files inside `data/` — the compose `osrm` service mounts `./data`.

---

## Contributing

- Open issues and PRs are welcome.
- Keep changes focused and small; add tests for new behavior in `internal/`.

---

## License

This repository does not include a license file. Add a `LICENSE` to make the intended usage explicit.

---

If you'd like, I can:

- add a short developer `Makefile` or convenience scripts to run the API + worker together
- generate a PNG/SVG diagram from the mermaid block and add it to the repo
- add a `LICENSE` file (which one do you prefer?)

Enjoy exploring DeliveryGO!
