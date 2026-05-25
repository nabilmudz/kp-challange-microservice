package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"order-service/internal/clients"
	"order-service/internal/infrastructure/events"
	redisClient "order-service/internal/infrastructure/redis"
	"order-service/internal/modules/orders/dto"
	"order-service/internal/modules/orders/entity"
	"order-service/internal/modules/orders/repository"
)

type OrderRepositoryInterface interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByProductID(ctx context.Context, productID string) ([]entity.Order, error)
	UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error
}

type ProductClientInterface interface {
	GetProduct(ctx context.Context, productID string) (*dto.ProductResponse, error)
}

type RedisInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

func NewOrderServiceWithDeps(
	repo OrderRepositoryInterface,
	productClient ProductClientInterface,
	redis RedisInterface,
) *OrderService {
	return &OrderService{
		repo:          repo,
		productClient: productClient,
		redis:         redis,
		publisher:     nil,
	}
}

var (
	ErrProductNotFound           = errors.New("PRODUCT_NOT_FOUND")
	ErrInsufficientStock         = errors.New("INSUFFICIENT_STOCK")
	ErrProductServiceUnavailable = errors.New("PRODUCT_SERVICE_UNAVAILABLE")
)

const orderCacheTTL = 30 * time.Second

type OrderService struct {
	repo          OrderRepositoryInterface
	productClient ProductClientInterface
	redis         RedisInterface
	publisher     *events.Publisher
}

func NewOrderService(
	repo *repository.OrderRepository,
	productClient *clients.ProductClient,
	redis *redisClient.RedisClient,
	publisher *events.Publisher,
) *OrderService {
	return &OrderService{repo: repo, productClient: productClient, redis: redis, publisher: publisher}
}

func cacheKey(productID string) string {
	return "orders:product:" + productID
}

func (s *OrderService) Create(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	productCacheKey := "product:" + req.ProductID
	var product *dto.ProductResponse

	if cached, err := s.redis.Get(ctx, productCacheKey); err == nil {
		product = &dto.ProductResponse{}
		if err := json.Unmarshal([]byte(cached), product); err != nil {
			product = nil
		} else {
			log.Printf("[Cache] HIT product:%s", req.ProductID)
		}
	}

	if product == nil {
		log.Printf("[Cache] MISS product:%s", req.ProductID)
		var err error
		product, err = s.productClient.GetProduct(ctx, req.ProductID)
		if err != nil {
			if errors.Is(err, clients.ErrProductNotFound) {
				return nil, ErrProductNotFound
			}
			if errors.Is(err, clients.ErrProductServiceUnavailable) {
				return nil, ErrProductServiceUnavailable
			}
			return nil, fmt.Errorf("failed to reach product-service: %w", err)
		}
		if data, err := json.Marshal(product); err == nil {
			_ = s.redis.Set(ctx, productCacheKey, string(data), 60*time.Second)
		}
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

	if s.publisher != nil {
		go func() {
			_ = s.publisher.PublishOrderCreated(context.Background(), events.OrderCreatedEvent{
				OrderID:   order.ID,
				ProductID: order.ProductID,
				Quantity:  order.Quantity,
				CreatedAt: order.CreatedAt,
			})
		}()
	}

	_ = s.redis.Del(ctx, cacheKey(req.ProductID))

	resp := dto.ToResponse(order)
	return &resp, nil
}

func (s *OrderService) GetByProductID(ctx context.Context, productID string) (*dto.OrderListResponse, error) {
	key := cacheKey(productID)
	if cached, err := s.redis.Get(ctx, key); err == nil {
		log.Printf("[Cache] HIT key=%s", key)
		var result dto.OrderListResponse
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return &result, nil
		}
	}
	log.Printf("[Cache] MISS key=%s", key)

	orders, err := s.repo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	responses := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		responses[i] = dto.ToResponse(&orders[i])
	}

	result := &dto.OrderListResponse{
		ProductID: productID,
		Total:     len(responses),
		Orders:    responses,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.redis.Set(ctx, key, data, orderCacheTTL)
	}

	return result, nil
}
