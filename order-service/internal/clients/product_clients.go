package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"order-service/internal/modules/orders/dto"
	"time"
)

var (
	ErrProductNotFound           = errors.New("PRODUCT_NOT_FOUND")
	ErrProductServiceUnavailable = errors.New("PRODUCT_SERVICE_UNAVAILABLE")
)

type ProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *ProductClient) GetProduct(ctx context.Context, productID string) (*dto.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/products/%s", c.baseURL, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("GetProduct: failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrProductServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProductNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrProductServiceUnavailable
	}

	var product dto.ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("GetProduct: failed to decode response: %w", err)
	}

	return &product, nil
}
