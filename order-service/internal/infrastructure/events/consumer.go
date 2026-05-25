package events

import (
	"encoding/json"
	"log"

	"order-service/internal/infrastructure/rabbitmq"
)

const (
	ExchangeProductCreated = "product.created"
)

type ProductCreatedEvent struct {
	ProductID string  `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
}

type Consumer struct {
	rmq *rabbitmq.RabbitMQ
}

func NewConsumer(rmq *rabbitmq.RabbitMQ) *Consumer {
	return &Consumer{rmq: rmq}
}

func (c *Consumer) ListenProductCreated() {
	if err := c.rmq.DeclareExchange(ExchangeProductCreated); err != nil {
		log.Fatalf("[Consumer] Failed to declare exchange: %v", err)
	}

	q, err := c.rmq.DeclareQueue("order-service.product.created")
	if err != nil {
		log.Fatalf("[Consumer] Failed to declare queue: %v", err)
	}

	if err := c.rmq.BindQueue(q.Name, ExchangeProductCreated); err != nil {
		log.Fatalf("[Consumer] Failed to bind queue: %v", err)
	}

	if err := c.rmq.Channel.Qos(1, 0, false); err != nil {
		log.Fatalf("[Consumer] Failed to set QoS: %v", err)
	}

	msgs, err := c.rmq.Channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[Consumer] Failed to consume: %v", err)
	}

	log.Println("[Consumer] Listening for product.created events...")

	go func() {
		for msg := range msgs {
			var event ProductCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("[Consumer] Failed to parse event: %v", err)
				msg.Nack(false, false)
				continue
			}
			log.Printf("[Consumer] product.created received productId=%s name=%s price=%.2f",
				event.ProductID, event.Name, event.Price)
			msg.Ack(false)
		}
	}()
}

func (c *Consumer) ListenOrderCreated() {
	if err := c.rmq.DeclareExchange(ExchangeOrderCreated); err != nil {
		log.Fatalf("[Consumer] Failed to declare exchange: %v", err)
	}

	q, err := c.rmq.DeclareQueue("order-service.order.created")
	if err != nil {
		log.Fatalf("[Consumer] Failed to declare queue: %v", err)
	}

	if err := c.rmq.BindQueue(q.Name, ExchangeOrderCreated); err != nil {
		log.Fatalf("[Consumer] Failed to bind queue: %v", err)
	}

	if err := c.rmq.Channel.Qos(1, 0, false); err != nil {
		log.Fatalf("[Consumer] Failed to set QoS: %v", err)
	}

	msgs, err := c.rmq.Channel.Consume(
		q.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("[Consumer] Failed to consume: %v", err)
	}

	log.Println("[Consumer] Listening for order.created events...")

	go func() {
		for msg := range msgs {
			var event OrderCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("[Consumer] Failed to parse event: %v", err)
				msg.Nack(false, false)
				continue
			}
			log.Printf("[Consumer] order.created received orderId=%s productId=%s quantity=%d",
				event.OrderID, event.ProductID, event.Quantity)
			msg.Ack(false)
		}
	}()
}
