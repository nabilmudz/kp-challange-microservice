package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"order-service/internal/infrastructure/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ExchangeOrderCreated = "order.created"

type OrderCreatedEvent struct {
	OrderID   string    `json:"orderId"`
	ProductID string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
}

type Publisher struct {
	rmq *rabbitmq.RabbitMQ
}

func NewPublisher(rmq *rabbitmq.RabbitMQ) *Publisher {
	p := &Publisher{rmq: rmq}

	if err := rmq.DeclareExchange(ExchangeOrderCreated); err != nil {
		log.Fatalf("[Publisher] Failed to declare exchange: %v", err)
	}

	return p
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.rmq.Channel.PublishWithContext(ctx,
		ExchangeOrderCreated,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("[Publisher] order.created published orderId=%s", event.OrderID)
	return nil
}
