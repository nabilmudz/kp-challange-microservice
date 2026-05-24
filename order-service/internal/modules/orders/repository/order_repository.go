package repository

import (
	"context"
	"fmt"

	"order-service/internal/modules/orders/entity"

	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *entity.Order) error {
	query := `
		INSERT INTO orders (id, product_id, quantity, total_price, status, created_at)
		VALUES (:id, :product_id, :quantity, :total_price, :status, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, order)
	return err
}

func (r *OrderRepository) FindByProductID(ctx context.Context, productID string) ([]entity.Order, error) {
	var orders []entity.Order
	query := `
		SELECT id, product_id, quantity, total_price, status, created_at
		FROM orders
		WHERE product_id = $1
		ORDER BY created_at DESC
	`
	if err := r.db.SelectContext(ctx, &orders, query, productID); err != nil {
		return nil, fmt.Errorf("FindByProductID: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
