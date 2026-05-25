package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"order-service/internal/modules/orders/dto"
	"order-service/internal/modules/orders/entity"
	"order-service/internal/modules/orders/service"

	"github.com/stretchr/testify/assert"
)

type mockOrderRepository struct {
	createFn        func(ctx context.Context, order *entity.Order) error
	findByProductFn func(ctx context.Context, productID string) ([]entity.Order, error)
}

func (m *mockOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	return m.createFn(ctx, order)
}
func (m *mockOrderRepository) FindByProductID(ctx context.Context, productID string) ([]entity.Order, error) {
	return m.findByProductFn(ctx, productID)
}
func (m *mockOrderRepository) UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error {
	return nil
}

type mockProductClient struct {
	getProductFn func(ctx context.Context, productID string) (*dto.ProductResponse, error)
}

func (m *mockProductClient) GetProduct(ctx context.Context, productID string) (*dto.ProductResponse, error) {
	return m.getProductFn(ctx, productID)
}

type mockRedis struct {
	store map[string]string
}

func newMockRedis() *mockRedis {
	return &mockRedis{store: make(map[string]string)}
}
func (m *mockRedis) Get(_ context.Context, key string) (string, error) {
	val, ok := m.store[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return val, nil
}
func (m *mockRedis) Set(_ context.Context, key string, value any, _ time.Duration) error {
	switch v := value.(type) {
	case string:
		m.store[key] = v
	case []byte:
		m.store[key] = string(v)
	default:
		return errors.New("unsupported type")
	}
	return nil
}
func (m *mockRedis) Del(_ context.Context, key string) error {
	delete(m.store, key)
	return nil
}

type mockPublisher struct{}

func (m *mockPublisher) PublishOrderCreated(_ context.Context, _ interface{}) error {
	return nil
}

func mockProduct() *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:       "product-uuid-123",
		Name:     "Mechanical Keyboard",
		Price:    850000,
		Quantity: 50,
	}
}

func mockOrder() entity.Order {
	return entity.Order{
		ID:         "order-uuid-123",
		ProductID:  "product-uuid-123",
		Quantity:   2,
		TotalPrice: 1700000,
		Status:     entity.StatusPending,
		CreatedAt:  time.Now(),
	}
}

func TestOrderService_Create_HappyPath(t *testing.T) {
	repo := &mockOrderRepository{
		createFn: func(_ context.Context, _ *entity.Order) error { return nil },
	}
	client := &mockProductClient{
		getProductFn: func(_ context.Context, _ string) (*dto.ProductResponse, error) {
			return mockProduct(), nil
		},
	}

	svc := service.NewOrderServiceWithDeps(repo, client, newMockRedis())
	result, err := svc.Create(context.Background(), &dto.CreateOrderRequest{
		ProductID: "product-uuid-123",
		Quantity:  2,
	})

	assert.NoError(t, err)
	assert.Equal(t, "product-uuid-123", result.ProductID)
	assert.Equal(t, 1700000.0, result.TotalPrice)
	assert.Equal(t, string(entity.StatusPending), result.Status)
}

func TestOrderService_GetByProductID_CacheHit(t *testing.T) {
	redis := newMockRedis()

	cached := &dto.OrderListResponse{
		ProductID: "product-uuid-123",
		Total:     1,
		Orders: []dto.OrderResponse{{
			ID:         "order-uuid-123",
			ProductID:  "product-uuid-123",
			Quantity:   2,
			TotalPrice: 1700000,
			Status:     "pending",
			CreatedAt:  time.Now().Format(time.RFC3339),
		}},
	}
	data, _ := json.Marshal(cached)
	redis.store["orders:product:product-uuid-123"] = string(data)

	repo := &mockOrderRepository{
		findByProductFn: func(_ context.Context, _ string) ([]entity.Order, error) {
			t.Error("repository should not be called on cache hit")
			return nil, nil
		},
	}

	svc := service.NewOrderServiceWithDeps(repo, &mockProductClient{}, redis)
	result, err := svc.GetByProductID(context.Background(), "product-uuid-123")

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}

func TestOrderService_GetByProductID_CacheMiss(t *testing.T) {
	redis := newMockRedis()
	order := mockOrder()

	repo := &mockOrderRepository{
		findByProductFn: func(_ context.Context, _ string) ([]entity.Order, error) {
			return []entity.Order{order}, nil
		},
	}

	svc := service.NewOrderServiceWithDeps(repo, &mockProductClient{}, redis)
	result, err := svc.GetByProductID(context.Background(), "product-uuid-123")

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Total)

	_, cacheErr := redis.Get(context.Background(), "orders:product:product-uuid-123")
	assert.NoError(t, cacheErr)
}
