package database

import (
	"fmt"
	"log"
	"time"

	"order-service/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgresDB(cfg *config.Config) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("[Postgres] Failed to connect: %v", err)
	}

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := migrate(db); err != nil {
		log.Fatalf("[Postgres] Migration failed: %v", err)
	}

	log.Println("[Postgres] Connected and migrated")
	return db
}

func migrate(db *sqlx.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		product_id  UUID NOT NULL,
		quantity    INT NOT NULL CHECK (quantity > 0),
		total_price DECIMAL(12, 2) NOT NULL,
		status      VARCHAR(20) NOT NULL DEFAULT 'pending',
		created_at  TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_orders_product_id ON orders(product_id);
	`
	_, err := db.Exec(query)
	return err
}
