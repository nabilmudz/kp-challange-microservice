package main

import (
	"log"
	"order-service/internal/clients"
	"order-service/internal/config"
	"order-service/internal/infrastructure/database"
	"order-service/internal/infrastructure/events"
	rmqInfra "order-service/internal/infrastructure/rabbitmq"
	redisInfra "order-service/internal/infrastructure/redis"
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
	redisClient := redisInfra.NewRedisClient(cfg)
	rmq := rmqInfra.NewRabbitMQ(cfg)
	defer rmq.Close()

	publisher := events.NewPublisher(rmq)
	consumer := events.NewConsumer(rmq)
	consumer.ListenProductCreated()
	consumer.ListenOrderCreated()

	orderSvc := service.NewOrderService(orderRepo, productClient, redisClient, publisher)

	orderHandler := handler.NewOrderHandler(orderSvc)

	router := gin.New()
	router.Use(gin.Recovery())

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
