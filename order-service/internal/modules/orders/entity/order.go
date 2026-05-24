package entity

import "time"

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusCompleted  OrderStatus = "completed"
	StatusFailed     OrderStatus = "failed"
)

type Order struct {
	ID         string      `db:"id"`
	ProductID  string      `db:"product_id"`
	Quantity   int         `db:"quantity"`
	TotalPrice float64     `db:"total_price"`
	Status     OrderStatus `db:"status"`
	CreatedAt  time.Time   `db:"created_at"`
}
