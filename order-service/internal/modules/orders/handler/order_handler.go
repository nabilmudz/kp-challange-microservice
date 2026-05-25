package handler

import (
	"errors"
	"net/http"

	"order-service/internal/modules/orders/dto"
	"order-service/internal/modules/orders/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.Create)
		orders.GET("/product/:productId", h.GetByProductID)
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      "Validation error",
			"message":    err.Error(),
		})
		return
	}

	result, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"statusCode": http.StatusNotFound,
				"error":      "Not found",
				"message":    "Product not found",
			})
		case errors.Is(err, service.ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{
				"statusCode": http.StatusConflict,
				"error":      "Insufficient stock",
				"message":    "Product quantity is not enough",
			})
		case errors.Is(err, service.ErrProductServiceUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"statusCode": http.StatusServiceUnavailable,
				"error":      "Service unavailable",
				"message":    "Product service is unavailable",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"error":      "Internal server error",
				"message":    err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusAccepted, result)
}

func (h *OrderHandler) GetByProductID(c *gin.Context) {
	productID := c.Param("productId")

	result, err := h.service.GetByProductID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "Internal server error",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
