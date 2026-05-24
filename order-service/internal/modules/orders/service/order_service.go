package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"order-service/internal/clients"
	"order-service/internal/modules/orders/dto"
	"order-service/internal/modules/orders/entity"
	"order-service/internal/modules/orders/repository"
)

var (
	ErrProductNotFound           = errors.New("PRODUCT_NOT_FOUND")
	ErrInsufficientStock         = errors.New("INSUFFICIENT_STOCK")
	ErrProductServiceUnavailable = errors.New("PRODUCT_SERVICE_UNAVAILABLE")
)

type OrderService struct {
	repo          *repository.OrderRepository
	productClient *clients.ProductClient
}

func NewOrderService(
	repo *repository.OrderRepository,
	productClient *clients.ProductClient,
) *OrderService {
	return &OrderService{
		repo:          repo,
		productClient: productClient,
	}
}

func (s *OrderService) Create(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	product, err := s.productClient.GetProduct(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, clients.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		if errors.Is(err, clients.ErrProductServiceUnavailable) {
			return nil, ErrProductServiceUnavailable
		}
		return nil, fmt.Errorf("failed to reach product-service: %w", err)
	}

	if product.Quantity < req.Quantity {
		return nil, ErrInsufficientStock
	}

	order := &entity.Order{
		ID:         uuid.New().String(),
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: product.Price * float64(req.Quantity),
		Status:     entity.StatusPending,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	resp := dto.ToResponse(order)
	return &resp, nil
}

func (s *OrderService) GetByProductID(ctx context.Context, productID string) (*dto.OrderListResponse, error) {
	orders, err := s.repo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	responses := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		responses[i] = dto.ToResponse(&orders[i])
	}

	return &dto.OrderListResponse{
		ProductID: productID,
		Total:     len(responses),
		Orders:    responses,
	}, nil
}
