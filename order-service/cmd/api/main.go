package main

import (
	"log"

	"order-service/internal/clients"
	"order-service/internal/config"
	"order-service/internal/infrastructure/database"
	"order-service/internal/modules/orders/handler"
	"order-service/internal/modules/orders/repository"
	"order-service/internal/modules/orders/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := database.NewPostgresDB(cfg)
	productClient := clients.NewProductClient(cfg.ProductSvc.URL)

	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo, productClient)
	orderHandler := handler.NewOrderHandler(orderSvc)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "order-service",
		})
	})

	orderHandler.RegisterRoutes(router)

	log.Printf("Starting Order Service on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
