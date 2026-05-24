package dto

import (
	"time"

	"order-service/internal/modules/orders/entity"
)

type CreateOrderRequest struct {
	ProductID string `json:"productId" binding:"required,uuid"`
	Quantity  int    `json:"quantity"       binding:"required,min=1"`
}

type OrderResponse struct {
	ID         string  `json:"id"`
	ProductID  string  `json:"productId"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"totalPrice"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
}

type OrderListResponse struct {
	ProductID string          `json:"productId"`
	Total     int             `json:"total"`
	Orders    []OrderResponse `json:"orders"`
}

type ProductResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"qty"`
}

func ToResponse(o *entity.Order) OrderResponse {
	return OrderResponse{
		ID:         o.ID,
		ProductID:  o.ProductID,
		Quantity:   o.Quantity,
		TotalPrice: o.TotalPrice,
		Status:     string(o.Status),
		CreatedAt:  o.CreatedAt.Format(time.RFC3339),
	}
}
