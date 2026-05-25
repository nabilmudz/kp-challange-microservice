# order-service

Golang service for managing orders.

## Stack

- Golang, Gin
- PostgreSQL
- Redis (caching)
- RabbitMQ (events)

## How to Run Locally

```bash
# start dependencies first
docker compose up order-db redis rabbitmq product-service -d

cp .env.example .env
go run ./cmd/api/main.go
```

## Endpoints

```bash
# Create order
curl -X POST http://localhost:3002/orders \
  -H "Content-Type: application/json" \
  -d '{"productId": "your-product-id", "quantity": 2}'

# Get orders by product ID
curl http://localhost:3002/orders/product/{productId}
```

## Caching

`GET /orders/product/:productId` is cached in Redis with a 30s TTL.  
Cache is invalidated when a new order is created for that product.

## Events

| Event | Type | Description |
|---|---|---|
| `order.created` | publish | Emitted after an order is created |
| `product.created` | consume | Logs incoming product info |

## Architecture

```
Handler → Service → Repository → PostgreSQL
              ↕
            Redis
              ↕
          RabbitMQ
```

## Notes

- Order creation validates product existence by calling product-service
- Designed to handle high throughput (tested with k6)