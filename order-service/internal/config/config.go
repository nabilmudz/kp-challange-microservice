package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	Database   DatabaseConfig
	Redis      RedisConfig
	RabbitMQ   RabbitMQConfig
	ProductSvc ProductSvcConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host string
	Port string
}

type RabbitMQConfig struct {
	URL string
}

type ProductSvcConfig struct {
	URL string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] .env not found, using environment variables")
	}

	return &Config{
		Port: getEnv("PORT", "3002"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "order_user"),
			Password: getEnv("DB_PASSWORD", "order_pass"),
			Name:     getEnv("DB_NAME", "order_db"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		},
		ProductSvc: ProductSvcConfig{
			URL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:3001"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
