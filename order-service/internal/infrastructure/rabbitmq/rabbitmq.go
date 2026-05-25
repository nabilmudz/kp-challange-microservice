package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"order-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(cfg *config.Config) *RabbitMQ {
	var conn *amqp.Connection
	var err error

	maxRetries := 5
	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(cfg.RabbitMQ.URL)
		if err == nil {
			break
		}
		delay := time.Duration(i) * time.Second
		log.Printf("[RabbitMQ] Connection failed, retry %d/%d in %s", i, maxRetries, delay)
		time.Sleep(delay)
	}
	if err != nil {
		log.Fatalf("[RabbitMQ] Failed to connect after %d retries: %v", maxRetries, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[RabbitMQ] Failed to open channel: %v", err)
	}

	conn.NotifyClose(make(chan *amqp.Error, 1))

	log.Println("[RabbitMQ] Connected")
	return &RabbitMQ{Conn: conn, Channel: ch}
}

func (r *RabbitMQ) DeclareExchange(name string) error {
	return r.Channel.ExchangeDeclare(
		name,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) DeclareQueue(name string) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) BindQueue(queueName, exchange string) error {
	return r.Channel.QueueBind(queueName, "", exchange, false, nil)
}

func (r *RabbitMQ) Close() {
	if err := r.Channel.Close(); err != nil {
		fmt.Printf("[RabbitMQ] Error closing channel: %v\n", err)
	}
	if err := r.Conn.Close(); err != nil {
		fmt.Printf("[RabbitMQ] Error closing connection: %v\n", err)
	}
}
