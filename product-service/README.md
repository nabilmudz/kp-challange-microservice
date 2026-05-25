# product-service

NestJS service for managing products.

## Stack

- NestJS, TypeScript
- PostgreSQL
- Redis (caching)
- RabbitMQ (events)

## How to Run Locally

```bash
# start dependencies first
docker compose up product-db redis rabbitmq -d

cp .env.example .env
npm install
npm run start:dev
```

## Endpoints

```bash
# Create product
curl -X POST http://localhost:3001/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Mechanical Keyboard", "price": 850000, "qty": 50}'

# Get product by ID
curl http://localhost:3001/products/{id}
```

## Caching

`GET /products/:id` is cached in Redis with a 60s TTL.  
Cache is invalidated when an `order.created` event is received.

## Events

| Event | Type | Description |
|---|---|---|
| `product.created` | publish | Emitted after a product is created |
| `order.created` | consume | Reduces product qty |

## Architecture

```
Controller → Service → Repository → PostgreSQL
                ↕
              Redis
                ↕
            RabbitMQ
```