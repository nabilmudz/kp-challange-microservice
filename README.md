# KP Challenge — Microservice App

A simple microservice application built with NestJS and Golang.

## Services

| Service | Stack | Port |
|---|---|---|
| product-service | NestJS + PostgreSQL | 3001 |
| order-service | Golang (Gin) + PostgreSQL | 3002 |

## Infrastructure

- **RabbitMQ** — event communication between services
- **Redis** — caching for product and order data
- **PostgreSQL** — separate database per service

## How to Run

```bash
cp .env.example .env
docker compose up -d
```

Services will be available at:
- Product: http://localhost:3001
- Order: http://localhost:3002
- RabbitMQ Management: http://localhost:15672

## Event Flow

```
POST /orders
    └── order-service publishes → order.created
            └── product-service consumes → reduces product qty

POST /products
    └── product-service publishes → product.created
            └── order-service consumes → logs product info
```

## Project Structure

```
.
├── product-service/   # NestJS
├── order-service/     # Golang
└── docker-compose.yml
```

## Load Test Results (k6)

- **Target**: 3000 rps for 30s
- **Environment**: local (Intel i5 Gen 12, 12GB RAM)
- **Actual throughput**: ~2977 rps
- **Error rate**: 0.42%
- **p95 latency**: 227ms
- **Median latency**: 64ms
- **p90 latency**: 176ms

All thresholds passed. Nearly 90k orders processed in 30 seconds.